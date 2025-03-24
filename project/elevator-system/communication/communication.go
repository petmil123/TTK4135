// Communication and management of state.
package communication

import (
	"Driver-go/elevator-system/assignerInterface"
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"time"
)

func RunCommunication(id string, numFloors int, port int, btnEvent chan elevio.ButtonEvent, orderComplete chan elevio.ButtonEvent, elevatorOrderCh chan state.ElevatorOrders, elevatorStateCh chan state.ElevatorState) {

	// Initialize state for ourselves
	orders := state.CreateStateStruct(id, numFloors)
	activePeers := make([]string, 1)
	activePeers[0] = id

	// Keep alive channels
	peerTxEnable := make(chan bool)

	// state channel
	peerUpdateCh := make(chan peers.PeerUpdate)

	go peers.Transmitter(21060, id, peerTxEnable)
	go peers.Receiver(21060, peerUpdateCh)

	stateTx := make(chan state.StateStruct)
	stateRx := make(chan state.StateStruct)

	go bcast.Transmitter(port, stateTx)
	go bcast.Receiver(port, stateRx)

	for {
		select {
		case <-time.After(50 * time.Millisecond):
			stateTx <- orders

		case receivedState := <-stateRx:
			orders.CompareIncoming(receivedState)
			elevatorOrderCh <- assignerInterface.AssignHallRequests(orders, activePeers)

		case peerUpdate := <-peerUpdateCh:
			activePeers = peerUpdate.Peers

		case buttonEvent := <-btnEvent:
			orders.SetButtonOrder(buttonEvent, true)
			elevatorOrderCh <- orders.GetConfirmedOrders(activePeers)
		case completedOrder := <-orderComplete:
			orders.SetButtonOrder(completedOrder, false)
			elevatorOrderCh <- orders.GetConfirmedOrders(activePeers)
		case elevatorState := <-elevatorStateCh:
			orders.SetElevatorState(elevatorState)
		}
	}
}

// Communication and management of state.
package communication

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
	"time"
)

func RunCommunication(id string, numFloors int, communicationPort int, peerPort int, btnEvent <-chan elevio.ButtonEvent, orderComplete <-chan elevio.ButtonEvent, assignerCh chan<- state.StateStruct, elevatorStateCh <-chan state.ElevatorState, txEnableCh chan bool) {

	// Initialize state for ourselves
	orders := state.CreateStateStruct(id, numFloors)
	activePeers := make([]string, 1)
	activePeers[0] = id

	// Keep alive channels
	peerTxEnable := make(chan bool)
	peerUpdateCh := make(chan peers.PeerUpdate)

	// state channel
	go peers.Transmitter(peerPort, id, peerTxEnable)
	go peers.Receiver(peerPort, peerUpdateCh)

	stateTx := make(chan state.StateStruct)
	stateRx := make(chan state.StateStruct)

	go bcast.Transmitter(communicationPort, stateTx)
	go bcast.Receiver(communicationPort, stateRx)

	for {
		select {

		case <-time.After(20 * time.Millisecond):
			stateTx <- orders

		case receivedState := <-stateRx:
			orders.CompareIncoming(receivedState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		case peerUpdate := <-peerUpdateCh:
			activePeers = peerUpdate.Peers
			if peerUpdate.New != "" {
				_, exists := orders.Orders[peerUpdate.New]
				fmt.Println("New peer: ", peerUpdate.New)
				if !exists {
					orders.Orders[peerUpdate.New] = state.CreateElevatorOrders(numFloors)
					orders.ElevatorStates[peerUpdate.New] = state.CreateElevatorState()
				}
			}
			if len(peerUpdate.Lost != 0) {
				fmt.Println("Lost peers: ", peerUpdate.Lost)
			}
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		case buttonEvent := <-btnEvent:
			orders.SetButtonOrder(buttonEvent, true)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case completedOrder := <-orderComplete:
			orders.SetButtonOrder(completedOrder, false)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case elevatorState := <-elevatorStateCh:
			orders.SetElevatorState(elevatorState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		case val := <-txEnableCh:
			peerTxEnable <- val
		}
	}
}

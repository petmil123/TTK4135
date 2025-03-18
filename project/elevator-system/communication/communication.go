package communication

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"time"
)

func RunCommunication(id string, numFloors int, port int, btnEvent chan elevio.ButtonEvent, orderCompleteOther chan state.OrderStruct, orderCompleteSelf chan elevio.ButtonEvent, newOrder chan state.OrderStruct) {

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
		case <-time.After(5000 * time.Millisecond):
			stateTx <- orders
		case receivedState := <-stateRx:
			orders.CompareIncoming(receivedState)
			orders.SendNewOrders(activePeers, newOrder, orderCompleteOther)

		case peerUpdate := <-peerUpdateCh:
			activePeers = peerUpdate.Peers

		case buttonEvent := <-btnEvent:
			orders.SetOrder(buttonEvent, true)
			orders.SendNewOrders(activePeers, newOrder, orderCompleteOther)

		case completedOrder := <-orderCompleteSelf:
			orders.SetOrder(completedOrder, false)
			orders.SendNewOrders(activePeers, newOrder, orderCompleteOther)
		}
	}
}

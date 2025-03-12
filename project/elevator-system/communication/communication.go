package communication

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"time"
)

func RunCommunication(id string, numFloors int, port int, btnEvent chan elevio.ButtonEvent, orderCompleteOther chan elevio.ButtonEvent, orderCompleteSelf chan elevio.ButtonEvent, newOrder chan elevio.ButtonEvent) {

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
			orders.SendNewHallOrders(activePeers, newOrder, orderCompleteOther)
			orders.SendNewCabOrders(activePeers, newOrder)

		case peerUpdate := <-peerUpdateCh:
			activePeers = peerUpdate.Peers

		case buttonEvent := <-btnEvent:
			orders.SetOrder(buttonEvent)
			if buttonEvent.Button == elevio.BT_Cab {
				orders.SendNewCabOrders(activePeers, newOrder)
			} else {
				orders.SendNewHallOrders(activePeers, newOrder, orderCompleteOther)
			}
		case completedOrder := <-orderCompleteSelf:
			orders.UnsetOrder(completedOrder)
			if completedOrder.Button == elevio.BT_Cab {
				orders.SendNewCabOrders(activePeers, newOrder)
			} else {
				orders.SendNewHallOrders(activePeers, newOrder, orderCompleteOther)
			}
		}
	}
}

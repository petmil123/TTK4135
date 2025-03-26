// Communication and management of state.
package communication

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"time"
)

// Channels:
// btnEventCh: channel for knowing whenever a button is pressed (incomming)
// orderCompleteCh: channel for knowing whenever a order is completed (incomming)
// assignerCh: channel for sending all elevator states (worldview) to the assigner (outgoing)
// elevatorStateCh: channel for getting the elevator states (incomming)

// RunCommunication handles network communication and state management from the state.go and elevio.go to the assigner.
func RunCommunication(id string, numFloors int, communicationPort int, peerPort int, btnEventCh <-chan elevio.ButtonEvent, orderCompleteCh <-chan elevio.ButtonEvent, assignerCh chan<- state.StateStruct, elevatorStateCh <-chan state.ElevatorState) {

	// Initialize state for ourselves
	orders := state.CreateStateStruct(id, numFloors)
	activePeers := make([]string, 1)
	activePeers[0] = id

	// Keep alive channels (heartbeats)
	peerTxEnable := make(chan bool)
	peerUpdateCh := make(chan peers.PeerUpdate)

	// state channel
	go peers.Transmitter(peerPort, id, peerTxEnable)
	go peers.Receiver(peerPort, peerUpdateCh)

	// state communication between elevators
	stateTx := make(chan state.StateStruct) //Sending our own state
	stateRx := make(chan state.StateStruct) // Getting the others state

	// broadcasting
	go bcast.Transmitter(communicationPort, stateTx)
	go bcast.Receiver(communicationPort, stateRx)

	for {
		select {

		// update rate
		case <-time.After(20 * time.Millisecond):
			stateTx <- orders

		// reciving state from the other elevators and update worldview
		case receivedState := <-stateRx:
			orders.CompareIncoming(receivedState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		// Update list of active elevators
		case peerUpdate := <-peerUpdateCh:
			activePeers = peerUpdate.Peers
			if peerUpdate.New != "" {
				_, exists := orders.Orders[peerUpdate.New]
				if !exists {
					orders.Orders[peerUpdate.New] = state.CreateElevatorOrders(numFloors)
					orders.ElevatorStates[peerUpdate.New] = state.CreateElevatorState()
				}

			}
			assignerCh <- orders.GetActivePeerWorldview(activePeers) // ? hvorfor står denne her, klarer ikke å forklare

		case ButtonEvent := <-btnEventCh:
			orders.SetButtonOrder(ButtonEvent, true)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case completedOrder := <-orderCompleteCh:
			orders.SetButtonOrder(completedOrder, false)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case elevatorState := <-elevatorStateCh:
			orders.SetElevatorState(elevatorState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		}
	}
}

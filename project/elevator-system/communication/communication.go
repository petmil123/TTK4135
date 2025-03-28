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

// Channels:
// btnEventCh: channel for knowing whenever a button is pressed (incomming)
// orderCompleteCh: channel for knowing whenever a order is completed (incomming)
// assignerCh: channel for sending all elevator states (worldview) to the assigner (outgoing)
// elevatorStateCh: channel for getting the elevator states (incomming)

// RunCommunication handles network communication and state management from the state.go and elevio.go to the assigner.
func RunCommunication(id string, numFloors int, communicationPort int, peerPort int, btnEvent <-chan elevio.ButtonEvent, orderComplete <-chan elevio.ButtonEvent, assignerCh chan<- state.StateStruct, elevatorStateCh <-chan state.ElevatorState, txEnableCh chan bool) {

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
			// Deep copy before sending
			toSend := state.StateStruct{
				Id:             orders.Id,
				ElevatorStates: make(map[string]state.ElevatorState),
				Orders:         make(map[string]state.ElevatorOrders),
			}
			for key, value := range orders.ElevatorStates {
				toSend.ElevatorStates[key] = value
			}
			for key, value := range orders.Orders {
				toSend.Orders[key] = value
			}

			stateTx <- toSend

		// reciving state from the other elevators and update worldview
		case receivedState := <-stateRx:
			orders.CompareIncoming(receivedState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		// Update list of active elevators
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

			if len(peerUpdate.Lost) != 0 {
				fmt.Println("Lost peers: ", peerUpdate.Lost)
			}
			fmt.Println("All known peers are now ", peerUpdate.Peers)
      // Keep the peer list of the assigner updated
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

		case ButtonEvent := <-btnEventCh:
			orders.SetButtonOrder(ButtonEvent, true)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case completedOrder := <-orderCompleteCh:
			orders.SetButtonOrder(completedOrder, false)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)
		case elevatorState := <-elevatorStateCh:
			orders.SetElevatorState(elevatorState)
			assignerCh <- orders.GetActivePeerWorldview(activePeers)

			//TODO: Remove and use channel directly
		case val := <-txEnableCh:
			peerTxEnable <- val
		}
	}
}

package communication

import (
	"Driver-go/elevator-system/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"time"
)

type CabCallStruct struct {
	Floor   int
	Active  bool
	AlterId uint8
}

type hallCallStruct struct {
	Floor   int
	Dir     int
	Active  bool
	AlterId uint8
}

type StateStruct struct {
	Id        string                      //id of the elevator sending
	CabCalls  map[string][]CabCallStruct  // Cab call for each elevator
	HallCalls map[string][]hallCallStruct //Hall calls as seen for each elevator
}

type ElevatorStateStruct struct {
	CabCalls  []CabCallStruct
	HallCalls []hallCallStruct
}

func RunCommunication(id string, port int, btnEvent chan elevio.ButtonEvent, orderComplete chan int, smCh chan ElevatorStateStruct) {

	// Initialize state for ourselves
	elevatorState := initializeState(id)

	// Keep alive channels
	peerTxEnable := make(chan bool)

	// state channel
	peerUpdateCh := make(chan peers.PeerUpdate)

	go peers.Transmitter(21060, id, peerTxEnable)
	go peers.Receiver(21060, peerUpdateCh)

	stateTx := make(chan StateStruct)
	stateRx := make(chan StateStruct)

	go bcast.Transmitter(port, stateTx)
	go bcast.Receiver(port, stateRx)

	for {
		select {
		case <-time.After(5000 * time.Millisecond):
			stateTx <- elevatorState
		case receivedState := <-stateRx:
			elevatorState := handleStateUpdate(elevatorState, receivedState) // This function has side effects
			smCh <- elevatorState

		case peerUpdate := <-peerUpdateCh:
			handlePeerUpdate(peerUpdate, elevatorState)

		case buttonEvent := <-btnEvent:
			elevatorState = handleButtonEvent(elevatorState, buttonEvent, id)
			// case floorCompleted := <-orderComgplete:
			// handleOrderComplete()
		}
	}
}
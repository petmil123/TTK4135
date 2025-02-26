package communication

import (
	"Driver-go/elevator-system/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"fmt"
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
	Id        string
	CabCalls  map[string][]CabCallStruct
	HallCalls []hallCallStruct
}

func RunCommunication(id string, port int, btnEvent chan elevio.ButtonEvent, orderComplete chan int) {
	// Store peers
	// activePeers := make([]string, 0)

	// Initialize state for ourselves
	state := initializeState(id)
	fmt.Println(state)
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
			stateTx <- state
		case receivedState := <-stateRx:
			state = handleStateUpdate(state, receivedState)
			fmt.Println(state)

			// case peerUpdate := <-peerUpdateCh:
			//	handlePeerUpdate()
		case buttonEvent := <-btnEvent:
			handleButtonEvent(state, buttonEvent, id)
			// case floorCompleted := <-orderComgplete:
			// handleOrderComplete()
		}
	}
}

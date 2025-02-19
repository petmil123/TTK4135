package communication

import (
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
	CabCalls  []CabCallStruct
	HallCalls []hallCallStruct
}

func handleStateUpdate(state *StateStruct, receivedState StateStruct) {
	fmt.Println("Received state update: ")
}

func RunCommunication(id string, port int) {
	// Store peers
	// activePeers := make([]string, 0)

	// Make dummy state to have something to send
	state := StateStruct{
		Id: id,
		CabCalls: []CabCallStruct{
			{Floor: 0, Active: false, AlterId: 0},
			{Floor: 1, Active: false, AlterId: 0},
			{Floor: 2, Active: false, AlterId: 0},
			{Floor: 3, Active: false, AlterId: 0},
		},
		HallCalls: []hallCallStruct{
			{Floor: 0, Dir: 0, Active: false, AlterId: 0},
			{Floor: 0, Dir: 1, Active: false, AlterId: 0},
			{Floor: 1, Dir: 0, Active: false, AlterId: 0},
			{Floor: 1, Dir: 1, Active: false, AlterId: 0},
			{Floor: 2, Dir: 0, Active: false, AlterId: 0},
			{Floor: 2, Dir: 1, Active: false, AlterId: 0},
			{Floor: 3, Dir: 0, Active: false, AlterId: 0},
			{Floor: 3, Dir: 1, Active: false, AlterId: 0},
		},
	}
	// Keep alive channels
	peerTxEnable := make(chan bool)

	// state channel
	peerUpdateCh := make(chan peers.PeerUpdate)

	go peers.Transmitter(20060, id, peerTxEnable)
	go peers.Receiver(20060, peerUpdateCh)

	stateTx := make(chan StateStruct)
	stateRx := make(chan StateStruct)

	go bcast.Transmitter(port, stateTx)
	go bcast.Receiver(port, stateRx)

	for {
		select {
		case <-time.After(1000 * time.Millisecond):
			stateTx <- state
		case receivedState := <-stateRx:
			handleStateUpdate(&state, receivedState)
		}
	}
}

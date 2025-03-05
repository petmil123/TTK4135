package communication

import (
	"Driver-go/elevator-system/elevio"
	"Network-go/network/peers"
)

func handleStateUpdate(state StateStruct, receivedState StateStruct) ElevatorStateStruct {
	id := receivedState.Id
	selfId := state.Id
	//Update incoming elevator
	state.CabCalls[id] = receivedState.CabCalls[id]
	state.HallCalls[id] = receivedState.HallCalls[id]

	//Check if the incoming state knows something you dont know (disregard peers you dont know about for now)
	for k, v := range receivedState.CabCalls {
		// Skip yourself
		if k == selfId {
			continue
		}
		// Skip any peers that are unknown, should we add these?
		if localCalls, ok := state.CabCalls[k]; ok {
			for i := 0; i < len(v); i++ {
				if v[i].AlterId > localCalls[i].AlterId {
					localCalls[i] = v[i]
				}
			}
			state.CabCalls[k] = localCalls
		}

	}

	for k, v := range receivedState.HallCalls {
		// Skip yourself
		if k == selfId {
			continue
		}
		// Skip any peers that are unknown, should we add these?
		if localCalls, ok := state.HallCalls[k]; ok {
			for i := 0; i < len(v); i++ {
				if v[i].AlterId > localCalls[i].AlterId {
					localCalls[i] = v[i]
				}
			}
			state.HallCalls[k] = localCalls
		}
	}
	return getElevatorState(state, selfId)
}

func initializeState(id string) StateStruct {

	// TODO: Remove hardcoded floor numbers
	cabCalls := make(map[string][]CabCallStruct)
	cabCalls[id] = []CabCallStruct{
		{Floor: 0, Active: false, AlterId: 0},
		{Floor: 1, Active: false, AlterId: 0},
		{Floor: 2, Active: false, AlterId: 0},
		{Floor: 3, Active: false, AlterId: 0},
	}

	hallCalls := make(map[string][]hallCallStruct)
	hallCalls[id] = []hallCallStruct{
		{Floor: 0, Dir: 0, Active: false, AlterId: 0},
		{Floor: 0, Dir: 1, Active: false, AlterId: 0},
		{Floor: 1, Dir: 0, Active: false, AlterId: 0},
		{Floor: 1, Dir: 1, Active: false, AlterId: 0},
		{Floor: 2, Dir: 0, Active: false, AlterId: 0},
		{Floor: 2, Dir: 1, Active: false, AlterId: 0},
		{Floor: 3, Dir: 0, Active: false, AlterId: 0},
		{Floor: 3, Dir: 1, Active: false, AlterId: 0},
	}

	return StateStruct{
		Id:        id,
		CabCalls:  cabCalls,
		HallCalls: hallCalls,
	}

}

func handleButtonEvent(state StateStruct, buttonPress elevio.ButtonEvent, id string) StateStruct {
	switch buttonPress.Button {
	case 0:
		state.HallCalls[id][buttonPress.Floor*2].Active = true
		state.HallCalls[id][buttonPress.Floor*2].AlterId += 1

	case 1:
		state.HallCalls[id][buttonPress.Floor*2+1].Active = true
		state.HallCalls[id][buttonPress.Floor*2+1].AlterId += 1

	case 2:
		state.CabCalls[id][buttonPress.Floor].Active = true
		state.CabCalls[id][buttonPress.Floor].AlterId += 1
	}

	return state
}

// handlePeerUpdates updates the structure of the state structure to accomodate for active peers
func handlePeerUpdate(p peers.PeerUpdate, globalState StateStruct) {
	// Add empty state for new peers
	for _, val := range p.New {
		peerId := string(val)

		globalState.CabCalls[peerId] = []CabCallStruct{
			{Floor: 0, Active: false, AlterId: 0},
			{Floor: 1, Active: false, AlterId: 0},
			{Floor: 2, Active: false, AlterId: 0},
			{Floor: 3, Active: false, AlterId: 0},
		}

		globalState.HallCalls[peerId] = []hallCallStruct{
			{Floor: 0, Dir: 0, Active: false, AlterId: 0},
			{Floor: 0, Dir: 1, Active: false, AlterId: 0},
			{Floor: 1, Dir: 0, Active: false, AlterId: 0},
			{Floor: 1, Dir: 1, Active: false, AlterId: 0},
			{Floor: 2, Dir: 0, Active: false, AlterId: 0},
			{Floor: 2, Dir: 1, Active: false, AlterId: 0},
			{Floor: 3, Dir: 0, Active: false, AlterId: 0},
			{Floor: 3, Dir: 1, Active: false, AlterId: 0},
		}
	}

	// remove lost peers
	for _, val := range p.Lost {
		peerId := string(val)

		delete(globalState.CabCalls, peerId)
		delete(globalState.HallCalls, peerId)
	}
}

// This is supposed to be more sophisticated
func getElevatorState(state StateStruct, id string) ElevatorStateStruct {
	var elevatorState ElevatorStateStruct
	elevatorState.CabCalls = state.CabCalls[id]
	elevatorState.HallCalls = state.HallCalls[id]

	return elevatorState
}

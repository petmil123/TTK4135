package communication

import "Driver-go/elevator-system/elevio"

func handleStateUpdate(state StateStruct, receivedState StateStruct) StateStruct {

	//Hall calls
	for i := 0; i < len(receivedState.HallCalls); i++ {
		if receivedState.HallCalls[i].AlterId > state.HallCalls[i].AlterId {
			state.HallCalls[i] = receivedState.HallCalls[i]
		}
	}
	//Cab calls
	for k, v := range receivedState.CabCalls {
		//If the sender has info on a peer we dont know about, add it
		if _, exists := state.CabCalls[k]; !exists {
			state.CabCalls[k] = receivedState.CabCalls[k]
		} else {
			// Compare cab calls
			stateCalls := state.CabCalls[k]
			for i := 0; i < len(v); i++ {
				if v[i].AlterId > stateCalls[i].AlterId {
					stateCalls[i] = v[i]
				}
			}
			state.CabCalls[k] = stateCalls
		}

	}
	return state
}

func handlePeerUpdate() {
	return
}

func initializeState(id string) StateStruct {

	cabCalls := make(map[string][]CabCallStruct)
	cabCalls[id] = []CabCallStruct{
		{Floor: 0, Active: false, AlterId: 0},
		{Floor: 1, Active: false, AlterId: 0},
		{Floor: 2, Active: false, AlterId: 0},
		{Floor: 3, Active: false, AlterId: 0},
	}

	return StateStruct{
		CabCalls: cabCalls,
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

}

func handleButtonEvent(state StateStruct, buttonPress elevio.ButtonEvent, id string) {
	switch buttonPress.Button {
	case 0:
		state.HallCalls[buttonPress.Floor*2].Active = true
		state.HallCalls[buttonPress.Floor*2].AlterId += 1

	case 1:
		state.HallCalls[buttonPress.Floor*2+1].Active = true
		state.HallCalls[buttonPress.Floor*2+1].AlterId += 1

	case 2:
		state.CabCalls[id][buttonPress.Floor].Active = true
		state.CabCalls[id][buttonPress.Floor].AlterId += 1
	}
}

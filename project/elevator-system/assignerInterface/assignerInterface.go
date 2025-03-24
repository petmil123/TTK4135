// // Interface for using the hall request assigner,
// // heavily inspired by the code here:
// // https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
package assignerInterface

import (
	"Driver-go/elevator-system/state"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func AssignHallRequests(worldview state.StateStruct, peerList []string) state.ElevatorOrders {
	executable := ""
	switch runtime.GOOS {
	case "linux":
		executable = "hall_request_assigner"
	case "windows":
		executable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
	jsonInput, err := json.Marshal(getHRAInput(worldview, peerList))
	if err != nil {
		fmt.Println("json.marshal error: ", err)
		return nil
	}

	ret, err := exec.Command("../"+executable, "-i", string(jsonInput)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}
	return getOrdersFromHRAOutput(worldview, *output, peerList[0])
}

func getHRAInput(worldview state.StateStruct, peerList []string) HRAInput {
	hraInput := HRAInput{}
	orders := worldview.GetConfirmedOrders(peerList)
	hallRequests := make([][2]bool, len(orders))
	for i, order := range orders {
		hallRequests[i] = [2]bool{order[0].Active, order[1].Active}
	}
	hraInput.HallRequests = hallRequests
	hraInput.States = make(map[string]HRAElevState)
	for _, id := range peerList {
		hraState := HRAElevState{}
		switch worldview.ElevatorStates[id].MachineState {
		case state.Idle:
			hraState.Behavior = "idle"
			hraState.Direction = "stop"
		case state.Up:
			hraState.Behavior = "moving"
			hraState.Direction = "up"
		case state.Down:
			hraState.Behavior = "moving"
			hraState.Direction = "down"
		case state.DoorOpen:
			hraState.Behavior = "doorOpen"
			hraState.Direction = "stop"
		}
		hraState.Floor = worldview.ElevatorStates[id].Floor
		hraState.CabRequests = make([]bool, len(worldview.Orders[id]))
		for i, order := range worldview.Orders[id] {
			hraState.CabRequests[i] = order[2].Active
		}
		hraInput.States[id] = hraState
	}
	return hraInput
}

func getOrdersFromHRAOutput(worldview state.StateStruct, output map[string][][2]bool, id string) state.ElevatorOrders {
	orders := state.CreateElevatorOrders(len(output[id]))
	for i, order := range output[id] {
		orders[i][0].Active = order[0]
		orders[i][1].Active = order[1]
		orders[i][2].Active = worldview.Orders[id][i][2].Active
	}
	return orders
}

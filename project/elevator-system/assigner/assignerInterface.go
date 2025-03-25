// // Interface for using the hall request assigner,
// // heavily inspired by the code here:
// // https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
package assigner

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

// Runs the hall request assigner and 
func AssignHallRequests(worldview state.StateStruct) state.ElevatorOrders {
	executable := ""
	switch runtime.GOOS {
	case "linux":
		executable = "hall_request_assigner"
	case "windows":
		executable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
	jsonInput, err := json.Marshal(getHRAInput(worldview))
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
	return getOrdersFromHRAOutput(worldview, *output)
}

// Takes the worldview of the known peers, and returns the data
// in the format that is necessary for the hall request assigner.
func getHRAInput(worldview state.StateStruct) HRAInput {
	// Get the confirmed hall requests. These are common for every elevator
	hraInput := HRAInput{}
	orders := worldview.GetConfirmedOrders()
	hallRequests := make([][2]bool, len(orders))
	for i, order := range orders {
		hallRequests[i] = [2]bool{order[0].Active, order[1].Active}
	}
	hraInput.HallRequests = hallRequests

	//Get elevator states and cab requests for each elevator.
	hraInput.States = make(map[string]HRAElevState)
	for key, elevator := range worldview.ElevatorStates {
		hraState := HRAElevState{}
		switch elevator.MachineState {
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
		hraState.Floor = elevator.Floor
		hraState.CabRequests = make([]bool, len(worldview.Orders[key]))
		for i, order := range worldview.Orders[key] {
			hraState.CabRequests[i] = order[2].Active
		}
		hraInput.States[key] = hraState
	}
	return hraInput
}

// Takes the HRA output and returns the orders for the elevator that runs the assigner.
func getOrdersFromHRAOutput(worldview state.StateStruct, output map[string][][2]bool) state.ElevatorOrders {
	id := worldview.Id
	orders := state.CreateElevatorOrders(len(output[id]))
	for i, order := range output[id] {
		orders[i][0].Active = order[0]
		orders[i][1].Active = order[1]
		orders[i][2].Active = worldview.Orders[id][i][2].Active
	}
	return orders
}

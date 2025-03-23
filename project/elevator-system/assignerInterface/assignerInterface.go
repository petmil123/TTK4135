// Interface for using the hall request assigner,
// most of this code is inspired by the code here:
// https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
package assignerInterface

import "runtime"

type AssignerInput struct {
	HallRequests [][2]bool
}

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

func AssignHallRequests(worlview ) {
	executable := ""
	switch runtime.GOOS {
	case "linux":
		executable = "hall_request_assigner"
	case "windows":
		executable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

}

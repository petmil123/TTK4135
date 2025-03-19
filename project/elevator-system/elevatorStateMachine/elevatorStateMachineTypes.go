package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"fmt"
	"time"
)

type StateMachineInputs struct {
	// Physical
	Obstruction  <-chan bool // Obstruction button toggled
	FloorArrival <-chan int  // Arrived at floor n
	//Communication
	StateCh <-chan state.ElevatorStateStruct //A new order is added to the orders
	//Internal
}

type StateMachineOutputs struct {
	OrderCompleted chan<- elevio.ButtonEvent
}

type MachineState int

const (
	Idle MachineState = iota
	Up
	Down
	DoorOpen
)

type ElevatorState struct {
	MachineState  MachineState              //Which state the FSM is in
	Obstructed    bool                      //Is the machine obstructed?
	PrevDirection elevio.MotorDirection     //Previous *moving* direction for choice of direction
	Orders        state.ElevatorStateStruct //Orders
	Floor         int                       //Which floor we previously were at.
	DoorTimer     *time.Timer               //Timer for the door
}

// Initialize the state of the elevator
func initializeElevator(numFloors int, timer *time.Timer) ElevatorState {
	return ElevatorState{
		MachineState:  Idle,
		Obstructed:    false,
		PrevDirection: elevio.MD_Stop,
		Orders:        state.CreateElevatorState(numFloors),
		Floor:         1,
		DoorTimer:     timer,
	}
}

// Sets the state after a state transition, and does everything that needs to be done with the elevator IO.
func (e *ElevatorState) setState(s MachineState) {
	switch s {
	case Idle:
		// Stop the engine
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(false)

		// Store previous state to keep it up to date.
		if e.MachineState == Up {
			e.PrevDirection = elevio.MD_Up
		} else if e.MachineState == Down {
			e.PrevDirection = elevio.MD_Down
		}
		e.MachineState = Idle
	case Up:
		elevio.SetMotorDirection(elevio.MD_Up)
		elevio.SetDoorOpenLamp(false)

		e.MachineState = Up

	case Down:
		elevio.SetMotorDirection(elevio.MD_Down)
		elevio.SetDoorOpenLamp(false)

		e.MachineState = Down

	case DoorOpen:
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		// Store previous state to keep it up to date.
		if e.MachineState == Up {
			e.PrevDirection = elevio.MD_Up
		} else if e.MachineState == Down {
			e.PrevDirection = elevio.MD_Down
		}
		e.DoorTimer.Reset(3 * time.Second)
		e.MachineState = DoorOpen
	}
	// fmt.Println(e.Orders)
}

// Sets the floor in state and handles IO
func (e *ElevatorState) setFloor(floor int) {
	e.Floor = floor
	elevio.SetFloorIndicator(floor)
}

// Checks if we have a cab call or a hall call in that direction
// for a given floor (if the direction is up or down)
func (e *ElevatorState) hasOrder(btn elevio.ButtonType, floor int) bool {
	return e.Orders[floor][btn].Active
}

func (e *ElevatorState) hasOrderAtFloor(floor int) bool {
	for _, order := range e.Orders[floor] {
		if order.Active {
			return true
		}
	}
	return false
}

// Checks if there is a cab order above the current floor
func (e *ElevatorState) hasCabOrderAbove(floor int) bool {
	for i := floor + 1; i < len(e.Orders); i++ {
		if e.Orders[i][2].Active {
			return true
		}
	}
	return false
}

// Checks if there is a cab order below the current floor
func (e *ElevatorState) hasCabOrderBelow(floor int) bool {
	for i := floor - 1; i >= 0; i-- {
		if e.Orders[i][2].Active {
			return true
		}
	}
	return false
}

// Checks if there is a hall order above the current floor
func (e *ElevatorState) hasHallOrderAbove(floor int) bool {
	for i := floor + 1; i < len(e.Orders); i++ {
		for j := 0; j < 2; j++ {
			if e.Orders[i][j].Active {
				return true
			}
		}
	}
	return false
}

// Checks if there is a hall order below the current floor
func (e *ElevatorState) hasHallOrderBelow(floor int) bool {
	for i := floor - 1; i >= 0; i-- {
		for j := 0; j < 2; j++ {
			if e.Orders[i][j].Active {
				return true
			}
		}
	}
	return false
}

// Clear orders when opening a door at the floor. We clear cab calls and hall calls
// in the direction of travel (should be future direction)
func (e *ElevatorState) clearOrders(s MachineState, floor int, outputCh chan<- elevio.ButtonEvent) {
	fmt.Println("Clearing orders")
	outputCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	if s == Up {
		outputCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
	} else if s == Down {
		outputCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
	}
}

func (e *ElevatorState) CalculateNextDir() MachineState {
	switch e.MachineState {
	case Up:
		if e.hasCabOrderAbove(e.Floor) {
			return Up
		} else if e.hasCabOrderBelow(e.Floor) {
			return Down
		} else if e.hasHallOrderAbove(e.Floor) {
			return Up
		} else if e.hasHallOrderBelow(e.Floor) {
			return Down
		} else {
			return Idle
		}
	case Down:
		if e.hasCabOrderBelow(e.Floor) {
			return Down
		} else if e.hasCabOrderAbove(e.Floor) {
			return Up
		} else if e.hasHallOrderBelow(e.Floor) {
			return Down
		} else if e.hasHallOrderAbove(e.Floor) {
			return Up
		} else {
			return Idle
		}
	}
	return Idle
}

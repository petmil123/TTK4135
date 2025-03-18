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
	OrderCompletedOther <-chan elevio.ButtonEvent //Another elevator finished this kind of order
	NewOrder            <-chan elevio.ButtonEvent //A new order is added to the orders
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

func (e *ElevatorState) setState(s MachineState) {
	//Sets the state after a state transition, and does everything that needs to be done with the elevator IO.
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
	fmt.Println(e.MachineState)
}

// Sets the floor in state and handles IO
func (e *ElevatorState) setFloor(floor int) {
	e.Floor = floor
	elevio.SetFloorIndicator(floor)
}

// Checks if we have a cab call or a hall call in that direction
// for a given floor (if the direction is up or down)
func (e *ElevatorState) hasOrders(s MachineState, floor int) bool {
	if e.Orders.CabCalls[floor].Active {
		return true
	} else if s == Up {
		return e.Orders.HallCalls[2*floor].Active
	} else if s == Down {
		return e.Orders.HallCalls[2*floor+1].Active
	} else {
		return e.Orders.HallCalls[2*floor].Active ||
			e.Orders.HallCalls[2*floor+1].Active
	}
}

func (e *ElevatorState) hasCabOrderAbove(floor int) bool {
	for i := floor + 1; i < len(e.Orders.CabCalls); i++ {
		if e.Orders.CabCalls[i].Active {
			return true
		}
	}
	return false
}

func (e *ElevatorState) hasCabOrderBelow(floor int) bool {
	for i := floor - 1; i >= 0; i-- {
		if e.Orders.CabCalls[i].Active {
			return true
		}
	}
	return false
}

func (e *ElevatorState) hasHallOrderAbove(floor int) bool {
	for i := floor + 1; i < len(e.Orders.CabCalls); i++ {
		if e.Orders.HallCalls[2*i].Active ||
			e.Orders.HallCalls[2*i+1].Active {
			return true
		}
	}
	return false
}

func (e *ElevatorState) hasHallOrderBelow(floor int) bool {
	for i := floor - 1; i >= 0; i-- {
		if e.Orders.HallCalls[2*i].Active ||
			e.Orders.HallCalls[2*i+1].Active {
			return true
		}
	}
	return false
}

func (e *ElevatorState) clearOrders(s MachineState, floor int) {
	if e.Orders.CabCalls[floor].Active {
		e.Orders.UnsetOrder(elevio.ButtonEvent{
			Floor:  floor,
			Button: elevio.BT_Cab,
		})
	}
	if s == Up {
		if e.Orders.HallCalls[2*floor].Active {
			e.Orders.UnsetOrder(elevio.ButtonEvent{
				Floor:  floor,
				Button: elevio.BT_HallUp,
			})
		}
	} else if s == Down {
		if e.Orders.HallCalls[2*floor+1].Active {
			e.Orders.UnsetOrder(elevio.ButtonEvent{
				Floor:  floor,
				Button: elevio.BT_HallDown,
			})
		}
	}
}

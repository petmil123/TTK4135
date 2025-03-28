package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"time"
)

type StateMachineInputs struct {
	// Physical
	Obstruction  <-chan bool // Obstruction button toggled
	FloorArrival <-chan int  // Arrived at floor n
	//Communication
	OrderCh <-chan state.ElevatorOrders //A new order is added to the orders
	//Internal
}

type StateMachineOutputs struct {
	OrderCompleted chan<- elevio.ButtonEvent
	StateCh        chan<- state.ElevatorState
	PeerTxEnableCh chan<- bool
}

type MachineState int

const (
	Idle MachineState = iota
	Up
	Down
	DoorOpen
)

type ElevatorState struct {
	MachineState    MachineState          //Which state the FSM is in
	Obstructed      bool                  //Is the machine obstructed?
	PrevDirection   elevio.MotorDirection //Previous *moving* direction for choice of direction
	NextDirection   MachineState          //The direction in which we have cleared a hall order
	Orders          state.ElevatorOrders  //Orders
	Floor           int                   //Which floor we previously were at.
	DoorTimer       *time.Timer           //Timer for the door
	StateErrorTimer *time.Timer           //Timer for noticing errors with the state
}

// Initialize the state of the elevator
func initializeElevator(numFloors int, doorTimer *time.Timer, stateErrorTimer *time.Timer) ElevatorState {
	elevio.SetMotorDirection(elevio.MD_Up)
	return ElevatorState{
		MachineState:    Up,
		Obstructed:      false,
		PrevDirection:   elevio.MD_Stop,
		NextDirection:   Idle,
		Orders:          state.CreateElevatorOrders(numFloors),
		Floor:           1,
		DoorTimer:       doorTimer,
		StateErrorTimer: stateErrorTimer,
	}
}

// Sets the state after a state transition, and does everything that needs to be done with the elevator IO.
func (e *ElevatorState) setState(newState MachineState) {
	if e.MachineState == DoorOpen {
		e.DoorTimer.Stop()
	}
	switch newState {
	case Idle:
		switch e.MachineState {
		case Up:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(false)
			e.StateErrorTimer.Stop()
			e.PrevDirection = elevio.MD_Up
			e.MachineState = Idle

		case Down:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(false)
			e.StateErrorTimer.Stop()
			e.PrevDirection = elevio.MD_Down
			e.MachineState = Idle
		case DoorOpen:
			elevio.SetDoorOpenLamp(false)
			e.StateErrorTimer.Stop()
			e.MachineState = Idle
		case Idle:
			// Do nothing
		}

	case Up:
		switch e.MachineState {
		case Up:
			e.StateErrorTimer.Reset(5 * time.Second)
		case Down:
			elevio.SetMotorDirection(elevio.MD_Up)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Up
		case DoorOpen:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Up)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Up
		case Idle:
			elevio.SetMotorDirection(elevio.MD_Up)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Up
		}

	case Down:
		switch e.MachineState {
		case Up:
			elevio.SetMotorDirection(elevio.MD_Down)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Down
		case Down:
			e.StateErrorTimer.Reset(5 * time.Second)
		case DoorOpen:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Down)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Down
		case Idle:
			elevio.SetMotorDirection(elevio.MD_Down)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.MachineState = Down
		}

	case DoorOpen:
		switch e.MachineState {
		case Up:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.DoorTimer.Reset(3 * time.Second)
			e.PrevDirection = elevio.MD_Up
			e.MachineState = DoorOpen
		case Down:
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.DoorTimer.Reset(3 * time.Second)
			e.PrevDirection = elevio.MD_Down
			e.MachineState = DoorOpen
		case DoorOpen:
			//Reset error timer if not obstructed (e.g. wrongful up call)
			if !e.Obstructed {
				e.StateErrorTimer.Reset(5 * time.Second)
			}
			e.DoorTimer.Reset(3 * time.Second)
		case Idle:
			elevio.SetDoorOpenLamp(true)
			e.StateErrorTimer.Reset(5 * time.Second)
			e.DoorTimer.Reset(3 * time.Second)
			e.MachineState = DoorOpen
		}
	}
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
//
// Also updates the nextDirection state.
func (e *ElevatorState) clearOrder(btn elevio.ButtonEvent, outputCh chan<- elevio.ButtonEvent) {
	if btn.Button == elevio.BT_HallUp {
		e.NextDirection = Up
	} else if btn.Button == elevio.BT_HallDown {
		e.NextDirection = Down
	}
	outputCh <- btn
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

func getCommState(e ElevatorState) state.ElevatorState {
	return state.ElevatorState{
		MachineState: state.MachineState(e.MachineState),
		Floor:        e.Floor,
	}
}

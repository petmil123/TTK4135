package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

// handleNewOrder processes a new button press event and updates the elevator state accordingly
func handleNewOrder(elevator *Elevator, order elevio.ButtonEvent, ch StateMachineChannels, doorTimedOut *time.Timer, engineErrorTimer *time.Timer) {
	// Add the new order to the elevator's queue and turn on the corresponding button lamp
	elevator.Queue[order.Floor][order.Button] = true
	elevio.SetButtonLamp(order.Button, order.Floor, true)

	switch elevator.State {
	case Idle:
		// If the elevator is idle, determine the direction based on the new order
		elevator.Dir = chooseDirection(*elevator)
		elevio.SetMotorDirection(elevator.Dir)
		if elevator.Dir == elevio.MD_Stop {
			// If no movement is needed, open the door at the current floor
			elevator.State = DoorOpen
			elevio.SetDoorOpenLamp(true)
			doorTimedOut.Reset(3 * time.Second)
			go func() { ch.OrderComplete <- order.Floor }()
			clearFloorQueue(elevator, elevator.Floor)
		} else {
			// Start moving towards the ordered floor
			elevator.State = Moving
			engineErrorTimer.Reset(3 * time.Second)
		}

	case DoorOpen:
		// If the door is already open and the new order is for the current floor, extend the door open time
		if elevator.Floor == order.Floor {
			doorTimedOut.Reset(3 * time.Second)
			go func() { ch.OrderComplete <- order.Floor }()
			clearFloorQueue(elevator, elevator.Floor)
		}
	}
	// Update the main controller with the new elevator state
	ch.Elevator <- *elevator
}

// handleArrivedAtFloor processes the event when the elevator arrives at a floor
func handleArrivedAtFloor(elevator *Elevator, ch StateMachineChannels, doorTimedOut *time.Timer, engineErrorTimer *time.Timer, orderCleared *bool) {
	// Set the floor indicator to the current floor
	elevio.SetFloorIndicator(elevator.Floor)
	fmt.Printf("Arrived at floor %d\n", elevator.Floor+1)

	if shouldStop(*elevator) {
		// If the elevator should stop at this floor, open the door and reset the timers
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevator.State = DoorOpen
		elevio.SetDoorOpenLamp(true)
		doorTimedOut.Reset(3 * time.Second)
		clearFloorQueue(elevator, elevator.Floor)
		go func() { ch.OrderComplete <- elevator.Floor }()
		engineErrorTimer.Stop()
	} else if elevator.State == Moving {
		// If the elevator is moving, reset the engine error timer
		engineErrorTimer.Reset(3 * time.Second)
	}
	// Update the main controller with the new elevator state
	ch.Elevator <- *elevator
}

// handleDoorTimeout processes the event when the door timer expires
func handleDoorTimeout(elevator *Elevator, ch StateMachineChannels, engineErrorTimer *time.Timer) {
	// Close the door and determine the next direction based on the queue
	elevio.SetDoorOpenLamp(false)
	elevator.Dir = chooseDirection(*elevator)
	if elevator.Dir == elevio.MD_Stop {
		// If no movement is needed, set the state to idle and stop the engine error timer
		elevator.State = Idle
		engineErrorTimer.Stop()
	} else {
		// Start moving towards the next ordered floor
		elevator.State = Moving
		engineErrorTimer.Reset(3 * time.Second)
		elevio.SetMotorDirection(elevator.Dir)
	}
	// Update the main controller with the new elevator state
	ch.Elevator <- *elevator
}

// handleEngineError processes the event when the engine error timer expires
func handleEngineError(elevator *Elevator, ch StateMachineChannels) {
	// Stop the motor to prevent any further movement
	elevio.SetMotorDirection(elevio.MD_Stop)

	// Set the elevator state to Undefined to indicate an error state
	elevator.State = Undefined

	// Print an error message to the console
	fmt.Println("Engine Error - System offline")

	// Send an error message to the StateError channel to notify the main controller
	ch.StateError <- fmt.Errorf("engine error detected")

	// Update the main controller with the new elevator state
	ch.Elevator <- *elevator
}

// handleObstruction processes the event when the obstruction switch is triggered or un-triggered
func handleObstruction(elevator *Elevator, obstruction bool, doorTimedOut *time.Timer) {
	elevator.Obstructed = obstruction
	if obstruction {
		// If the door is obstructed, reset the door timer to keep the door open
		doorTimedOut.Reset(3 * time.Second)
	} else if elevator.State == DoorOpen {
		// If the obstruction is cleared and the door is open, allow the door to close after the timer expires
		doorTimedOut.Reset(3 * time.Second)
	}
}

// chooseDirection determines the next direction of the elevator based on the queue
func chooseDirection(elevator Elevator) elevio.MotorDirection {
	switch elevator.Dir {
	case elevio.MD_Stop:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		}
	case elevio.MD_Up:
		if ordersAbove(elevator) {
			return elevio.MD_Up
		} else if ordersBelow(elevator) {
			return elevio.MD_Down
		}
	case elevio.MD_Down:
		if ordersBelow(elevator) {
			return elevio.MD_Down
		} else if ordersAbove(elevator) {
			return elevio.MD_Up
		}
	}
	return elevio.MD_Stop
}

// shouldStop determines if the elevator should stop at the current floor
func shouldStop(elevator Elevator) bool {
	if elevator.State != Moving {
		return false
	}

	switch elevator.Dir {
	case elevio.MD_Up:
		return elevator.Queue[elevator.Floor][elevio.BT_HallUp] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersAbove(elevator)
	case elevio.MD_Down:
		return elevator.Queue[elevator.Floor][elevio.BT_HallDown] ||
			elevator.Queue[elevator.Floor][elevio.BT_Cab] ||
			!ordersBelow(elevator)
	}
	return true
}

// ordersAbove checks if there are any orders above the current floor
func ordersAbove(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < len(elevator.Queue); f++ {
		for b := 0; b < 3; b++ {
			if elevator.Queue[f][b] {
				return true
			}
		}
	}
	return false
}

// ordersBelow checks if there are any orders below the current floor , b<3 because there are only 3 buttons
func ordersBelow(elevator Elevator) bool {
	for f := 0; f < elevator.Floor; f++ {
		for b := 0; b < 3; b++ {
			if elevator.Queue[f][b] {
				return true
			}
		}
	}
	return false
}

// clearFloorQueue clears all orders for the specified floor
func clearFloorQueue(elevator *Elevator, floor int) {
	for b := 0; b < 3; b++ {
		elevator.Queue[floor][b] = false
		elevio.SetButtonLamp(elevio.ButtonType(b), floor, false)
	}
}

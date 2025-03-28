// FSM that runs the elevator. Receives orders, and handles them accordingly.
// Channels:
//
// ObstructionCh: Gives status of the obstruction switch
//
// FloorArrivalCh: Notifies when we arrive at a floor.
//
// OrderCh: Notifies us of new orders that need to be handled.
//
// OrderCompletedCh: Channel to notify when we have cleared an order
//
// StateCh: Channel to notify of changes in elevator state
//
// PeerTxEnableCh: Channel to (de)activate heartbeats in failure cases.
package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

// Runs the elevator state machine and handles incoming orders.
func RunElevator(inputs StateMachineInputs, outputs StateMachineOutputs, numFloors int) {
	//Timer to keep the door open
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()

	// Timer to observe when we are stuck due to obstruction or motor power loss
	stateErrorTimer := time.NewTimer(5 * time.Second)
	stateErrorTimer.Stop()

	elevator := initializeElevator(numFloors, doorTimer, stateErrorTimer)
	for {
		select {
		case obstructed := <-inputs.ObstructionCh:
			elevator.Obstructed = obstructed
			switch elevator.MachineState {
			case DoorOpen:
				//Resets the timer if door is obstructed and open
				if obstructed {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
				} else {
					stateErrorTimer.Stop()
					outputs.PeerTxEnableCh <- true
				}
			}

		// Procedure for elevator behavior when arriving a floor
		case arrivedFloor := <-inputs.FloorArrivalCh:
			elevator.setFloor(arrivedFloor)
			fmt.Println("arrived at floor", elevator.Floor)
			switch elevator.MachineState {
			case Up:
				outputs.PeerTxEnableCh <- true
				// Decide which orders to clear
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompletedCh)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompletedCh)
					// Edge case
					if !elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) && elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompletedCh)
					}
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompletedCh)
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				} else {
					elevator.setState(Idle)
					outputs.StateCh <- getCommState(elevator)
				}
			case Down:
				outputs.PeerTxEnableCh <- true
				// Decide which orders to clear
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompletedCh)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompletedCh)
					// Edge case
					if !elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) && elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompletedCh)
					}
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompletedCh)
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else {
					elevator.setState(Idle)
					outputs.StateCh <- getCommState(elevator)
				}
			}

		// update the elevator order and go to next assigned order
		case receivedState := <-inputs.OrderCh:
			elevator.Orders = receivedState
			switch elevator.MachineState {
			case Idle:
				if elevator.hasOrderAtFloor(elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompletedCh)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompletedCh)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompletedCh)
				} else if elevator.hasCabOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasCabOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				}
			case DoorOpen:

			}

		// decide what to do then doorTimer runs out
		case <-elevator.DoorTimer.C:
			fmt.Println("Door timer")
			if elevator.Obstructed {
				fmt.Println("Obstruction, keep on in door open.")
				elevator.setState(DoorOpen)
			} else {
				switch elevator.NextDirection {
				case Up, Idle:
					if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
						elevator.setState(Up)
						outputs.StateCh <- getCommState(elevator)
					} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompletedCh)
						elevator.setState(DoorOpen)
						outputs.StateCh <- getCommState(elevator)
					} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.NextDirection = Down
						elevator.setState(DoorOpen)
						outputs.StateCh <- getCommState(elevator)
					} else {
						elevator.setState(Idle)
						outputs.StateCh <- getCommState(elevator)
					}
				case Down:
					if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.setState(Down)
					} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompletedCh)
						elevator.setState(DoorOpen)
						outputs.StateCh <- getCommState(elevator)
					} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.NextDirection = Up
						elevator.setState(DoorOpen)
						outputs.StateCh <- getCommState(elevator)
					} else {
						elevator.setState(Idle)
						outputs.StateCh <- getCommState(elevator)
					}
				}
			}
		// Handle being stuck.
		case <-elevator.StateErrorTimer.C:
			fmt.Println("State error timer")
			switch elevator.MachineState {
			case Up, Down:
				fmt.Println("State error timer")
				outputs.PeerTxEnableCh <- false
			case DoorOpen:
				if elevator.Obstructed {
					fmt.Println("State error timer")
					outputs.PeerTxEnableCh <- false
				}

			}
		}
	}
}

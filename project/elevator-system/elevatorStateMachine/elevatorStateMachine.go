package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

func RunElevator(inputs StateMachineInputs, outputs StateMachineOutputs, numFloors int) {
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	stateErrorTimer := time.NewTimer(5 * time.Second)
	stateErrorTimer.Stop()

	elevator := initializeElevator(numFloors, doorTimer, stateErrorTimer)
	for {
		select {
		case obstructed := <-inputs.Obstruction:
			elevator.Obstructed = obstructed
			switch elevator.MachineState {
			case DoorOpen:
				//Reset timer if it is turned on.
				if obstructed {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
				} else {
					stateErrorTimer.Stop()
					outputs.PeerTxEnableCh <- true
				}
			}

		case arrivedFloor := <-inputs.FloorArrival:
			elevator.setFloor(arrivedFloor)
			fmt.Println("arrived at floor", elevator.Floor)
			switch elevator.MachineState {
			case Up:
				outputs.PeerTxEnableCh <- true
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompleted)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompleted)
					if !elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) && elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
					}
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				} else {
					elevator.setState(Idle)
					outputs.StateCh <- getCommState(elevator)
				}
			case Down:
				outputs.PeerTxEnableCh <- true
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompleted)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
					outputs.StateCh <- getCommState(elevator)
				} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompleted)
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
					outputs.StateCh <- getCommState(elevator)
				} else {
					elevator.setState(Idle)
					outputs.StateCh <- getCommState(elevator)
				}
			}

		case receivedState := <-inputs.OrderCh:
			elevator.Orders = receivedState
			switch elevator.MachineState {
			case Idle:
				if elevator.hasOrderAtFloor(elevator.Floor) {
					elevator.setState(DoorOpen)
					outputs.StateCh <- getCommState(elevator)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompleted)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
					elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}, outputs.OrderCompleted)
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
					} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) { //TODO: Må vi ha cab her?
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
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
					} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) { //TODO: Må vi ha cab her?
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompleted)
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

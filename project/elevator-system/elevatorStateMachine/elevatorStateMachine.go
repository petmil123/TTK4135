package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
	"fmt"
	"time"
)

func RunElevator(inputs StateMachineInputs, outputs StateMachineOutputs, numFloors int) {
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	elevator := initializeElevator(numFloors, doorTimer)
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
				}
			}

		case arrivedFloor := <-inputs.FloorArrival:
			fmt.Println("arrived at floor", elevator.Floor)
			elevator.setFloor(arrivedFloor)
			switch elevator.MachineState {
			case Up:
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
			for floor, floorOrder := range receivedState {
				for btn, order := range floorOrder {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, order.Active)
				}
			}
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
				} else {
					elevator.setState(Idle)
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
					} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) { //TODO: Må vi ha cab her?
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}, outputs.OrderCompleted)
						elevator.setState(DoorOpen)
					} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.NextDirection = Down
						elevator.setState(DoorOpen)
					} else {
						elevator.setState(Idle)
					}
				case Down:
					if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.setState(Down)
					} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) { //TODO: Må vi ha cab her?
						elevator.clearOrder(elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}, outputs.OrderCompleted)
						elevator.setState(DoorOpen)
					} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
						elevator.NextDirection = Up
						elevator.setState(DoorOpen)
					} else {
						elevator.setState(Idle)
					}
				}
			}
		}
	}
}

func getCommState(e ElevatorState) state.ElevatorState {
	return state.ElevatorState{
		MachineState: state.MachineState(e.MachineState),
		Floor:        e.Floor,
		AlterId:      0,
	}
}

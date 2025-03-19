package elevatorStateMachine

import (
	"Driver-go/elevator-system/elevio"
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
				}
			}

		case arrivedFloor := <-inputs.FloorArrival:
			elevator.setFloor(arrivedFloor)
			switch elevator.MachineState {
			case Up:
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.setState(DoorOpen)
					elevator.clearOrder()
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else if elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else {
					elevator.setState(Idle)
				}
			case Down:
				if elevator.hasOrder(elevio.BT_Cab, elevator.Floor) || elevator.hasOrder(elevio.BT_HallDown, elevator.Floor) {
					elevator.setState(DoorOpen)
				} else if elevator.hasCabOrderBelow(elevator.Floor) || elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else if elevator.hasOrder(elevio.BT_HallUp, elevator.Floor) {
					elevator.setState(DoorOpen)
				} else if elevator.hasCabOrderAbove(elevator.Floor) || elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else {
					elevator.setState(Idle)
				}
			}

		case receivedState := <-inputs.StateCh:
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
					elevator.clearOrders(elevator.MachineState, elevator.Floor, outputs.OrderCompleted)
				} else if elevator.hasCabOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else if elevator.hasCabOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else if elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else if elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else {
					elevator.setState(Idle)
				}
			}

		case <-elevator.DoorTimer.C:
			fmt.Println("Door timer")
			if elevator.Obstructed {
				fmt.Println("Obstruction, keep on in door open.")
				elevator.setState(DoorOpen)
			} else {
				if elevator.hasOrders(elevator.MachineState, elevator.Floor) {
					elevator.setState(DoorOpen)
					elevator.clearOrders(elevator.MachineState, elevator.Floor, outputs.OrderCompleted)
				} else if elevator.hasCabOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else if elevator.hasCabOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else if elevator.hasHallOrderBelow(elevator.Floor) {
					elevator.setState(Down)
				} else if elevator.hasHallOrderAbove(elevator.Floor) {
					elevator.setState(Up)
				} else {
					elevator.setState(Idle)
				}
			}
		}
	}
}

package elevatorStateMachine

import (
	"fmt"
	"time"
)

func RunElevator(inputs StateMachineInputs, outputs StateMachineOutputs, numFloors int) {
	doorTimer := time.NewTimer(3 * time.Second)
	elevator := initializeElevator(numFloors, doorTimer)
	for {
		select {
		case obstructed := <-inputs.Obstruction:
			elevator.Obstructed = obstructed
			switch elevator.MachineState {
			case DoorOpen:
				//Effectively resets the timer
				elevator.setState(DoorOpen)
			}
		case arrivedFloor := <-inputs.FloorArrival:
			elevator.setFloor(arrivedFloor)
			switch elevator.MachineState {
			case Up:
				if elevator.hasOrders(elevator.MachineState, arrivedFloor) {
					elevator.setState(DoorOpen)
					elevator.clearOrders(elevator.MachineState, arrivedFloor)
				} else if elevator.hasCabOrderAbove(arrivedFloor) {
					elevator.setState(Up)
				} else if elevator.hasCabOrderBelow(arrivedFloor) {
					elevator.setState(Down)
				} else if elevator.hasHallOrderAbove(arrivedFloor) {
					elevator.setState(Up)
				} else if elevator.hasHallOrderBelow(arrivedFloor) {
					elevator.setState(Down)
				} else {
					elevator.setState(Idle)
				}
			case Down:
				if elevator.hasOrders(elevator.MachineState, arrivedFloor) {
					elevator.setState(DoorOpen)
					elevator.clearOrders(elevator.MachineState, arrivedFloor)
				} else if elevator.hasCabOrderBelow(arrivedFloor) {
					elevator.setState(Down)
				} else if elevator.hasCabOrderAbove(arrivedFloor) {
					elevator.setState(Up)
				} else if elevator.hasHallOrderBelow(arrivedFloor) {
					elevator.setState(Down)
				} else if elevator.hasHallOrderAbove(arrivedFloor) {
					elevator.setState(Up)
				} else {
					elevator.setState(Idle)
				}
			}
		case completedOrder := <-inputs.OrderCompletedOther:
			elevator.Orders.SetOrder(completedOrder.Order, false)
		case newOrder := <-inputs.NewOrder:
			elevator.Orders.SetOrder(newOrder.Order, true)
			switch elevator.MachineState {
			case Idle:
				if newOrder.Order.Floor == elevator.Floor {
					elevator.setState(DoorOpen)
				} else if newOrder.Order.Floor > elevator.Floor {
					elevator.setState(Up)
				} else {
					elevator.setState(Down)
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
					elevator.clearOrders(elevator.MachineState, elevator.Floor)
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

// // StateMachineChannels contains channels for communication between
// // elevator state machine, hardware interface, and main controller
// type StateMachineChannels struct {
// 	OrderComplete      chan int                       // Signals when an order is completed at a floor
// 	Elevator           chan Elevator                  // Updates main controller with elevator state
// 	StateError         chan error                     // Reports errors in elevator operation
// 	NewOrder           chan elevio.ButtonEvent        // Receives new button press events
// 	ArrivedAtFloor     chan int                       // Signals when elevator arrives at a floor
// 	Obstruction        chan bool                      // Signals when obstruction switch is activated
// 	Orders             chan state.ElevatorStateStruct // Orders that need to be handled
// 	OrderCompleteOther chan elevio.ButtonEvent        //Orders completed at another elevator
// }

// // Elevator represents the elevator state and properties
// type Elevator struct {
// 	State      ElevStateMachineState // Current state (Idle, Moving, etc)
// 	Dir        elevio.MotorDirection // Current direction of movement
// 	Floor      int                   // Current floor position
// 	Queue      [][]bool              // Order queue matrix [floors][button_types]
// 	Obstructed bool                  // True if obstruction switch is activated
// }

// // State represents the different states the elevator can be in
// type ElevStateMachineState int

// const (
// 	Idle      ElevStateMachineState = iota // Elevator is stationary with no orders
// 	Moving                                 // Elevator is moving between floors
// 	DoorOpen                               // Elevator is stopped with door open
// 	Undefined                              // Error state
// )

// // RunElevator initializes and runs the elevator state machine
// func RunElevator(ch StateMachineChannels, numFloors int) {
// 	// Initialize elevator with default values
// 	elevator := Elevator{
// 		State: Idle,
// 		Dir:   elevio.MD_Stop,
// 		Floor: elevio.GetFloor(),
// 		Queue: make([][]bool, numFloors), // Creates queue matrix based on number of floors
// 	}

// 	// Initialize queue for all floor buttons
// 	for i := range elevator.Queue {
// 		elevator.Queue[i] = make([]bool, 3) // BT_HallUp, BT_HallDown, BT_Cab
// 	}

// 	doorTimedOut := time.NewTimer(3 * time.Second)
// 	engineErrorTimer := time.NewTimer(3 * time.Second)
// 	doorTimedOut.Stop()
// 	engineErrorTimer.Stop()
// 	orderCleared := false
// 	ch.Elevator <- elevator

// 	for {
// 		select {
// 		case newOrder := <-ch.NewOrder:
// 			fmt.Println("NEW ORDER!")
// 			handleNewOrder(&elevator, newOrder, ch, doorTimedOut, engineErrorTimer)

// 		case elevator.Floor = <-ch.ArrivedAtFloor:
// 			handleArrivedAtFloor(&elevator, ch, doorTimedOut, engineErrorTimer, &orderCleared)

// 		case <-doorTimedOut.C:
// 			handleDoorTimeout(&elevator, ch, engineErrorTimer, doorTimedOut)

// 		case <-engineErrorTimer.C:
// 			handleEngineError(&elevator, ch)

// 		case obstruction := <-ch.Obstruction:
// 			handleObstruction(&elevator, obstruction, doorTimedOut)
// 		case orders := <-ch.Orders:
// 			fmt.Println("Orders received from comm module:")
// 			fmt.Println(orders)
// 			handleNetworkOrders(&elevator, orders, ch)
// 		}

// 	}
// }

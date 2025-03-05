package elevatorStateMachine

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevio"
	"fmt"
	"time"
)

// StateMachineChannels contains channels for communication between
// elevator state machine, hardware interface, and main controller
type StateMachineChannels struct {
	OrderComplete  chan int                // Signals when an order is completed at a floor
	Elevator       chan Elevator           // Updates main controller with elevator state
	StateError     chan error              // Reports errors in elevator operation
	NewOrder       chan elevio.ButtonEvent // Receives new button press events
	ArrivedAtFloor chan int                // Signals when elevator arrives at a floor
	Obstruction    chan bool               // Signals when obstruction switch is activated
	Orders         chan communication.ElevatorStateStruct
}

// Elevator represents the elevator state and properties
type Elevator struct {
	State      State                 // Current state (Idle, Moving, etc)
	Dir        elevio.MotorDirection // Current direction of movement
	Floor      int                   // Current floor position
	Queue      [][]bool              // Order queue matrix [floors][button_types]
	Obstructed bool                  // True if obstruction switch is activated
}

// State represents the different states the elevator can be in
type State int

const (
	Idle      State = iota // Elevator is stationary with no orders
	Moving                 // Elevator is moving between floors
	DoorOpen               // Elevator is stopped with door open
	Undefined              // Error state
)

// RunElevator initializes and runs the elevator state machine
func RunElevator(ch StateMachineChannels, numFloors int) {
	// Initialize elevator with default values
	elevator := Elevator{
		State: Idle,
		Dir:   elevio.MD_Stop,
		Floor: elevio.GetFloor(),
		Queue: make([][]bool, numFloors), // Creates queue matrix based on number of floors
	}

	// Initialize queue for all floor buttons
	for i := range elevator.Queue {
		elevator.Queue[i] = make([]bool, 3) // BT_HallUp, BT_HallDown, BT_Cab
	}

	doorTimedOut := time.NewTimer(3 * time.Second)
	engineErrorTimer := time.NewTimer(3 * time.Second)
	doorTimedOut.Stop()
	engineErrorTimer.Stop()
	orderCleared := false
	ch.Elevator <- elevator

	for {
		select {
		case newOrder := <-ch.NewOrder:
			fmt.Println("NEW ORDER!")
			handleNewOrder(&elevator, newOrder, ch, doorTimedOut, engineErrorTimer)

		case elevator.Floor = <-ch.ArrivedAtFloor:
			handleArrivedAtFloor(&elevator, ch, doorTimedOut, engineErrorTimer, &orderCleared)

		case <-doorTimedOut.C:
			handleDoorTimeout(&elevator, ch, engineErrorTimer, doorTimedOut)

		case <-engineErrorTimer.C:
			handleEngineError(&elevator, ch)

		case obstruction := <-ch.Obstruction:
			handleObstruction(&elevator, obstruction, doorTimedOut)
		case orders := <-ch.Orders:
			fmt.Println("Orders received from comm module:")
			fmt.Println(orders)
			handleNetworkOrders(&elevator, orders, ch)
		}

	}
}

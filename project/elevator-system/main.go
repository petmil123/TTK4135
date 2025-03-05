package main

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"flag"
	"fmt"
)

func main() {

	// Get id from command line argument
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")

	var port string
	flag.StringVar(&port, "port", "15657", "port of elevator")

	var numFloors int
	flag.IntVar(&numFloors, "numFloors", 4, "Number of floors in elevator system")
	flag.Parse()

	ip := "localhost:"

	elevio.Init(ip+port, numFloors)

	// Create channels for state machine
	ch := elevatorStateMachine.StateMachineChannels{
		OrderComplete:  make(chan int),
		Elevator:       make(chan elevatorStateMachine.Elevator),
		StateError:     make(chan error),
		NewOrder:       make(chan elevio.ButtonEvent),
		ArrivedAtFloor: make(chan int),
		Obstruction:    make(chan bool),
		Orders:         make(chan communication.ElevatorStateStruct),
	}

	// Start the elevator state machine
	go elevatorStateMachine.RunElevator(ch, numFloors)

	// Create driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Keep alive channels

	// Start polling
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go communication.RunCommunication(id, 20060, drv_buttons, ch.OrderComplete, ch.Orders)

	// Start network

	for {
		select {
		// case buttonEvent := <-drv_buttons:
		// fmt.Printf("Button pressed: %+v\n", buttonEvent)
		// ch.NewOrder <- buttonEvent

		case floor := <-drv_floors:
			fmt.Printf("Floor sensor: %+v\n", floor)
			ch.ArrivedAtFloor <- floor

		case obstruction := <-drv_obstr:
			fmt.Printf("Obstruction: %+v\n", obstruction)
			ch.Obstruction <- obstruction

		case stop := <-drv_stop:
			fmt.Printf("Stop button: %+v\n", stop)
			if stop {
				// Clear all button lamps
				// for f := 0; f < numFloors; f++ {
				// 	for b := elevio.ButtonType(0); b < 3; b++ {
				// 		elevio.SetButtonLamp(b, f, false)
				// 	}
				// }
				fmt.Println("Stop button pressed")
			}

		case elevator := <-ch.Elevator:
			fmt.Printf("Elevator state updated: %+v\n", elevator)

		case floor := <-ch.OrderComplete:
			fmt.Printf("Order completed at floor: %d\n", floor)

		case err := <-ch.StateError:
			fmt.Printf("Error: %v\n", err)

		}
	}
}

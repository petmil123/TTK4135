package main

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"flag"
)

func main() {

	// Get id from command line argument
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")

	var elevioPort string
	flag.StringVar(&elevioPort, "port", "15657", "port of elevator server")

	var numFloors int
	flag.IntVar(&numFloors, "numFloors", 4, "Number of floors in elevator system")
	flag.Parse()

	ip := "localhost:"

	elevio.Init(ip+elevioPort, numFloors)

	// Create driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	//Channels for passing orders between state machine and communication.
	orderCompletedSelf := make(chan elevio.ButtonEvent)
	orderCompletedOther := make(chan elevio.ButtonEvent)
	newOrder := make(chan elevio.ButtonEvent)

	// Keep alive channels

	// Start polling
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	go elevatorStateMachine.RunElevator(elevatorStateMachine.StateMachineInputs{
		Obstruction:         drv_obstr,
		FloorArrival:        drv_floors,
		OrderCompletedOther: orderCompletedOther,
		NewOrder:            newOrder,
	}, elevatorStateMachine.StateMachineOutputs{
		OrderCompleted: orderCompletedSelf,
	}, numFloors)
	go communication.RunCommunication(id, numFloors, 20060, drv_buttons, orderCompletedOther, orderCompletedSelf, newOrder)
	select {}
	// Start network
	// for {
	// 	select {
	// 	// case buttonEvent := <-drv_buttons:
	// 	// fmt.Printf("Button pressed: %+v\n", buttonEvent)
	// 	// ch.NewOrder <- buttonEvent

	// 	case floor := <-drv_floors:
	// 		fmt.Printf("Floor sensor: %+v\n", floor)
	// 		ch.ArrivedAtFloor <- floor

	// 	case obstruction := <-drv_obstr:
	// 		fmt.Printf("Obstruction: %+v\n", obstruction)
	// 		ch.Obstruction <- obstruction

	// 	case stop := <-drv_stop:
	// 		fmt.Printf("Stop button: %+v\n", stop)
	// 		if stop {
	// 			// Clear all button lamps
	// 			// for f := 0; f < numFloors; f++ {
	// 			// 	for b := elevio.ButtonType(0); b < 3; b++ {
	// 			// 		elevio.SetButtonLamp(b, f, false)
	// 			// 	}
	// 			// }
	// 			fmt.Println("Stop button pressed")
	// 		}

	// 	case elevator := <-ch.Elevator:
	// 		fmt.Printf("Elevator state updated: %+v\n", elevator)

	// 	case floor := <-ch.OrderComplete:
	// 		fmt.Printf("Order completed at floor: %d\n", floor)

	// 	case err := <-ch.StateError:
	// 		fmt.Printf("Error: %v\n", err)

	// 	}
	// }
}

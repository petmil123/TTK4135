package main

import (
	"Driver-go/elevator-system/communication"
	"Driver-go/elevator-system/elevatorStateMachine"
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
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
	orderCompletedSelf := make(chan elevio.ButtonEvent, 4) //Added buffer to not block
	orderCh := make(chan state.ElevatorOrders)
	stateCh := make(chan state.ElevatorState, 4)
	// Keep alive channels

	// Start polling
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	go elevatorStateMachine.RunElevator(elevatorStateMachine.StateMachineInputs{
		Obstruction:  drv_obstr,
		FloorArrival: drv_floors,
		OrderCh:      orderCh,
	}, elevatorStateMachine.StateMachineOutputs{
		OrderCompleted: orderCompletedSelf,
		StateCh:        stateCh,
	}, numFloors)
	go communication.RunCommunication(id, numFloors, 20060, drv_buttons, orderCompletedSelf, orderCh, stateCh)

	select {}

}

package main

import (
	"Driver-go/elevator-system/assigner"
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

	var elevatorServerPort string
	flag.StringVar(&elevatorServerPort, "elevatorPort", "15657", "port of elevator server")

	var elevatorServerIp string
	flag.StringVar(&elevatorServerIp, "elevatorIp", "localhost", "IP address of elevator server")

	var communicationPort int
	flag.IntVar(&communicationPort, "communicationPort", 20060, "Port for main communication")

	var peerPort int
	flag.IntVar(&peerPort, "peerPort", 21060, "Port for peer keep-alive communication")

	var numFloors int
	flag.IntVar(&numFloors, "numFloors", 4, "Number of floors in elevator system")
	flag.Parse()

	elevio.Init(elevatorServerIp+":"+elevatorServerPort, numFloors)

	// Create driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	//Channels for passing orders between state machine and communication.
	orderCompletedSelf := make(chan elevio.ButtonEvent, 4) //Added buffer to not block
	orderCh := make(chan state.ElevatorOrders, 4)
	stateCh := make(chan state.ElevatorState, 4)
	worldviewCh := make(chan state.StateStruct, 4)

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
	go communication.RunCommunication(id, numFloors, communicationPort, peerPort, drv_buttons, orderCompletedSelf, worldviewCh, stateCh)
	go assigner.RunAssigner(worldviewCh, orderCh, numFloors)

	select {}

}

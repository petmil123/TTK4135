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
	flag.StringVar(&id, "id", "1", "id of this peer")

	// Get port and IP adress from server
	var elevatorServerPort string
	flag.StringVar(&elevatorServerPort, "elevatorPort", "15657", "port of elevator server")

	var elevatorServerIp string
	flag.StringVar(&elevatorServerIp, "elevatorIp", "localhost", "IP address of elevator server")

	// Set communication ports
	var communicationPort int
	flag.IntVar(&communicationPort, "communicationPort", 20060, "Port for main communication")

	var peerPort int
	flag.IntVar(&peerPort, "peerPort", 21060, "Port for peer keep-alive communication")

	// Set number of floors (default = 4)
	var numFloors int
	flag.IntVar(&numFloors, "numFloors", 4, "Number of floors in elevator system")
	flag.Parse()

	// Initialize driver
	elevio.Init(elevatorServerIp+":"+elevatorServerPort, numFloors)

	// Create driver channels (Incomming from driver)
	drv_buttonsCh := make(chan elevio.ButtonEvent) //Channel for Button events
	drv_floorsCh := make(chan int)                 // Channel for floors. Sends when elevator arrives at new floor
	drv_obstrCh := make(chan bool)                 // Channel for obstruction. Sends True if obstruction occurs

	//Channels for passing orders between state machine and communication.
	orderCompletedSelfCh := make(chan elevio.ButtonEvent, 4) // Channel for notifying when an order is complete
	orderCh := make(chan state.ElevatorOrders, 4)            // Channel for getting orders assigned to an elevator ??
	stateCh := make(chan state.ElevatorState, 4)             // Channel for sending the elevator state
	worldviewCh := make(chan state.StateStruct, 64)          // Channel for updating all the elevator's state (worldview)

	// Start polling
	go elevio.PollButtons(drv_buttonsCh)
	go elevio.PollFloorSensor(drv_floorsCh)
	go elevio.PollObstructionSwitch(drv_obstrCh)

	// Start Elevator State Machine
	go elevatorStateMachine.RunElevator(elevatorStateMachine.StateMachineInputs{
		Obstruction:  drv_obstrCh,
		FloorArrival: drv_floorsCh,
		OrderCh:      orderCh,
	}, elevatorStateMachine.StateMachineOutputs{
		OrderCompleted: orderCompletedSelfCh,
		StateCh:        stateCh,
	}, numFloors)

	// Start network communication
	go communication.RunCommunication(id, numFloors, communicationPort, peerPort, drv_buttonsCh, orderCompletedSelfCh, worldviewCh, stateCh)

	// Start the order assiger
	go assigner.RunAssigner(worldviewCh, orderCh, numFloors)

	select {}

}

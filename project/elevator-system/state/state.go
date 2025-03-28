package state

import (
	"Driver-go/elevator-system/elevio"
	"fmt"
)

// Contains the status of a single order
type OrderStruct struct {
	Order   elevio.ButtonEvent //Identifier
	Active  bool               //Is the order active?
	AlterId uint               //Cyclic counter
}

// Type for the calls of each elevator
type ElevatorOrders [][]OrderStruct // A 2D slice where rows represent floors and columns represent button types

// The possible states of one elevator
type MachineState int

const (
	Idle MachineState = iota
	Up
	Down
	DoorOpen
)

// The elevator state needed for the HRA to make its decisions.
type ElevatorState struct {
	MachineState MachineState //Which state the FSM is in
	Floor        int          //Which floor we previously were at.
	AlterId      uint         //Cyclic counter
}

// Struct for the Worldview
type StateStruct struct {
	Id             string                    //id of the elevator sending
	ElevatorStates map[string]ElevatorState  //The state of the elevator
	Orders         map[string]ElevatorOrders // Map of all seen elevators
}

// "Constructor" for ElevatorState with all orders inactive. Returns the order structure for an elevator
func CreateElevatorOrders(numFloors int) ElevatorOrders {

	calls := make([][]OrderStruct, numFloors)
	for i := 0; i < numFloors; i++ {
		floorCalls := make([]OrderStruct, 3)
		for j := 0; j < 3; j++ {
			floorCalls[j].Order = elevio.ButtonEvent{
				Floor:  i,
				Button: elevio.ButtonType(j),
			}
			floorCalls[j].Active = false
			floorCalls[j].AlterId = 1
		}
		calls[i] = floorCalls
	}
	return calls
}

// Initialize the elevator state
func CreateElevatorState() ElevatorState {
	return ElevatorState{
		MachineState: Idle,
		Floor:        0,
		AlterId:      1,
	}
}

// Initialize global worldview
func CreateStateStruct(id string, numFloors int) StateStruct {
	elevatorOrders := make(map[string]ElevatorOrders)
	elevatorOrders[id] = CreateElevatorOrders(numFloors)
	elevatorStates := make(map[string]ElevatorState)
	elevatorStates[id] = CreateElevatorState()

	return StateStruct{
		Id:             id,
		Orders:         elevatorOrders,
		ElevatorStates: elevatorStates,
	}
}

// Compares all orders for a single elevator and updates if there are more recent edits.
func (own *ElevatorOrders) compareIncoming(incoming ElevatorOrders) {
	for i, incomingFloors := range incoming {
		for j, incomingOrder := range incomingFloors {
			if incomingOrder.AlterId == ^uint(0) && (*own)[i][j].AlterId == 0 {
				continue
			} else if incomingOrder.AlterId > (*own)[i][j].AlterId {
				(*own)[i][j] = incomingOrder
			} else if incomingOrder.AlterId == 0 && (*own)[i][j].AlterId == ^uint(0) {
				(*own)[i][j] = incomingOrder
			}
		}
	}
}

// Compares hall calls and updates for more recent edits.
func (own *ElevatorOrders) compareIncomingHall(incoming ElevatorOrders) {
	for i, incomingFloors := range incoming {
		for j := 0; j < 2; j++ {
			if incomingFloors[j].AlterId == ^uint(0) && (*own)[i][j].AlterId == 0 {
				continue
			} else if incomingFloors[j].AlterId > (*own)[i][j].AlterId {
				(*own)[i][j] = incomingFloors[j]
			} else if incomingFloors[j].AlterId == 0 && (*own)[i][j].AlterId == ^uint(0) {
				(*own)[i][j] = incomingFloors[j]
			}
		}
	}
}

// Compares the alterIDs and find out who has the newest one
func (own *ElevatorState) compareIncoming(incoming ElevatorState) {
	if incoming.AlterId == ^uint(0) && own.AlterId == 0 {

	} else if incoming.AlterId > own.AlterId {
		own.MachineState = incoming.MachineState
		own.AlterId = incoming.AlterId
		own.Floor = incoming.Floor
	} else if incoming.AlterId == 0 && own.AlterId == ^uint(0) {
		own.MachineState = incoming.MachineState
		own.AlterId = incoming.AlterId
		own.Floor = incoming.Floor
	}
}

// Looks at incoming state and updates state based on alterId.
func (own *StateStruct) CompareIncoming(incoming StateStruct) {

	// Orders:
	//For each elevator incoming knows about, update old info and add potential not known about elevators
	for key, incoming_val := range incoming.Orders {
		own_val, exists := own.Orders[key]
		if exists {
			own_val.compareIncoming(incoming_val)
			own.Orders[key] = own_val
		} else {
			//Add it to your state without comparing
			own.Orders[key] = incoming_val
		}
	}

	//Also, update hall calls so that new hall calls from incoming are propagated to ourselves
	//TODO: Make it so that if the elevator is idle or something like that, the order is not sent to the rest of the elevators.
	//Problem then is if there is a fault, who takes it? Is it a mess to reassign?
	own_val, exists := own.Orders[own.Id] //Does not work without this check     ?????
	if exists {
		own_val.compareIncomingHall(incoming.Orders[incoming.Id])
		own.Orders[own.Id] = own_val
	}

	// Elevator states:
	for key, incoming_val := range incoming.ElevatorStates {
		own_val, exists := own.ElevatorStates[key]
		if exists {
			own_val.compareIncoming(incoming_val)
			own.ElevatorStates[key] = own_val
		} else {
			own.ElevatorStates[key] = incoming_val
		}
	}
}

// This is the function for button presses and clearing orders.
func (elev *ElevatorOrders) SetButtonOrder(btn elevio.ButtonEvent, val bool) {
	if (*elev)[btn.Floor][btn.Button].Active != val {
		(*elev)[btn.Floor][btn.Button].Active = val
		if (*elev)[btn.Floor][btn.Button].AlterId == ^uint(0) {
			(*elev)[btn.Floor][btn.Button].AlterId = 0
		} else {
			(*elev)[btn.Floor][btn.Button].AlterId++
		}
	}
}

// Sets an order at itself in the worldview state. ?
func (s *StateStruct) SetButtonOrder(btn elevio.ButtonEvent, val bool) {
	elevator, exists := s.Orders[s.Id]
	if exists {
		elevator.SetButtonOrder(btn, val)
		s.Orders[s.Id] = elevator
	} else {
		panic("Elevator state does not know about itself!")
	}
}

// Updates the elevator state
func (s *StateStruct) SetElevatorState(state ElevatorState) {
	elevator, exists := s.ElevatorStates[s.Id]
	if exists {
		elevator.setState(state)
		s.ElevatorStates[s.Id] = elevator
	} else {
		panic("Elevator state does not know about itself!")
	}
}

// Function setting a new state
func (elev *ElevatorState) setState(state ElevatorState) {
	elev.MachineState = state.MachineState
	elev.Floor = state.Floor
	if elev.AlterId == ^uint(0) {
		elev.AlterId = 0
	} else {
		elev.AlterId++
	}
}

// Gets the orders that all peers agree on.
func (s *StateStruct) GetConfirmedOrders(numFloors int) ElevatorOrders {
	ownOrders := s.Orders[s.Id]
	toReturn := CreateElevatorOrders(numFloors) //ensures correct dimensions
	//TODO: Finn bedre ut av dette
	if len(s.Orders) == 0 {
		fmt.Println("Nå var du heldig")
		return toReturn
	}
	for floor, floorOrders := range ownOrders {
		for btn, orders := range floorOrders { // For each order
			minId := orders
			for _, peerOrders := range s.Orders { // For each peer
				if peerOrders[floor][btn].AlterId < minId.AlterId {
					minId = peerOrders[floor][btn]
				}
			}
			toReturn[floor][btn] = minId
		}
	}
	return toReturn
}

// Extracts the worldview state for active peers
func (s *StateStruct) GetActivePeerWorldview(peerList []string) StateStruct {
	toReturn := StateStruct{
		Id:             s.Id,
		Orders:         make(map[string]ElevatorOrders),
		ElevatorStates: make(map[string]ElevatorState),
	}

	for _, p := range peerList {
		toReturn.ElevatorStates[p] = s.ElevatorStates[p]
		toReturn.Orders[p] = s.Orders[p]
	}
	//Always include own state to make lights work
	toReturn.ElevatorStates[s.Id] = s.ElevatorStates[s.Id]
	toReturn.Orders[s.Id] = s.Orders[s.Id]

	return toReturn
}

// Debug utility:)
func (s *StateStruct) Prettyprint() {
	fmt.Println("Elevator id: ", s.Id)
	for key, elevator := range s.ElevatorStates {
		fmt.Println("Machine state: ", elevator.MachineState)
		fmt.Println("Floor: ", elevator.Floor)
		fmt.Println("Alter id: ", elevator.AlterId)
		fmt.Println("Orders: ")
		for floor, floorOrders := range s.Orders[key] {
			fmt.Println("Floor ", floor, ": hall down: ", floorOrders[0].Active, "(alter )", floorOrders[0].AlterId,
				": hall up: ", floorOrders[1].Active, "(alter )", floorOrders[1].AlterId,
				": cab: ", floorOrders[2].Active, "(alter )", floorOrders[2].AlterId)
		}

	}

}

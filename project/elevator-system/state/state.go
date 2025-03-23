package state

import (
	"Driver-go/elevator-system/elevio"
)

// Contains the status of a single order
type OrderStruct struct {
	Order   elevio.ButtonEvent //Identifier
	Active  bool               //Is the order active?
	AlterId uint8              //Cyclic counter
}

// Type for the calls of each elevator
type ElevatorOrders [][]OrderStruct //Should we specify how many calls there are at each floor?

type ElevatorState 

// Struct for the Worldview
type StateStruct struct {
	Id     string                    //id of the elevator sending
	Orders map[string]ElevatorOrders // Map of all seen elevators
}

// "Constructor" for ElevatorState with all orders inactive.
func CreateElevatorState(numFloors int) ElevatorOrders {

	calls := make([][]OrderStruct, numFloors)
	for i := 0; i < numFloors; i++ {
		floorCalls := make([]OrderStruct, 3)
		for j := 0; j < 3; j++ {
			floorCalls[j].Order = elevio.ButtonEvent{
				Floor:  i,
				Button: elevio.ButtonType(j),
			}
			floorCalls[j].Active = false
			floorCalls[j].AlterId = 0
		}
		calls[i] = floorCalls
	}
	return calls
}

// Initialize global worldview
func CreateStateStruct(id string, numFloors int) StateStruct {
	elevators := make(map[string]ElevatorOrders)
	elevators[id] = CreateElevatorState(numFloors)

	return StateStruct{
		Id:     id,
		Orders: elevators,
	}
}

// Compares all orders for a single elevator and updates if there are more recent edits.
func (own *ElevatorOrders) compareIncoming(incoming ElevatorOrders) {
	for i, incomingFloors := range incoming {
		for j, incomingOrder := range incomingFloors {
			if incomingOrder.AlterId > (*own)[i][j].AlterId {
				(*own)[i][j] = incomingOrder
			}
		}
	}
}

// Compares hall calls and updates for more recent edits.
func (own *ElevatorOrders) compareIncomingHall(incoming ElevatorOrders) {
	for i, incomingFloors := range incoming {
		for j := 0; j < 2; j++ {
			if incomingFloors[j].AlterId > (*own)[i][j].AlterId {
				(*own)[i][j] = incomingFloors[j]
			}
		}
	}
}

// Looks at incoming state and updates state based on alterId.
func (own *StateStruct) CompareIncoming(incoming StateStruct) {

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
	own_val, exists := own.Orders[own.Id] //Does not work without this check
	if exists {
		own_val.compareIncomingHall(incoming.Orders[incoming.Id])
		own.Orders[own.Id] = own_val
	}
}

// Compares single order sent and updates if newer, for message passing.
func (own *ElevatorOrders) CompareIncomingSingle(incoming OrderStruct) {
	btn := incoming.Order
	if (*own)[btn.Floor][btn.Button].AlterId <= incoming.AlterId {
		(*own)[btn.Floor][btn.Button] = incoming
	}
}

// This is the function for button presses and clearing orders.
func (elev *ElevatorOrders) SetButtonOrder(btn elevio.ButtonEvent, val bool) {
	if (*elev)[btn.Floor][btn.Button].Active != val {
		(*elev)[btn.Floor][btn.Button].Active = val
		(*elev)[btn.Floor][btn.Button].AlterId++
	}
}

// Sets an order at itself in the worldview state.
func (s *StateStruct) SetButtonOrder(btn elevio.ButtonEvent, val bool) {
	elevator, exists := s.Orders[s.Id]
	if exists {
		elevator.SetButtonOrder(btn, val)
		s.Orders[s.Id] = elevator
	} else {
		panic("Elevator state does not know about itself!")
	}
}

// Check if all peers know about a hall call order, and sends a message on a
// channel if this is the case.
//
// Not 100% sure this is the right way to do it, some issues might occur
func (s *StateStruct) SendNewOrders(peerList []string,
	newOrderCh chan<- OrderStruct, completedOrderCh chan<- OrderStruct) {
	for floor, floorCalls := range s.Orders[s.Id] {
		for i, call := range floorCalls {
			allEqual := true
			for _, p := range peerList {
				if s.Orders[p][floor][i].Active != call.Active {
					allEqual = false
				}
			}
			if allEqual {
				if floorCalls[i].Active {
					newOrderCh <- floorCalls[i]
				} else if i < 2 { // Disregard cab calls
					completedOrderCh <- floorCalls[i]
				}
			}
		}

	}
}

func (s *StateStruct) GetConfirmedOrders(peerList []string) ElevatorOrders {
	self := s.Orders[s.Id]
	toReturn := CreateElevatorState(len(self))
	for floor, floorOrders := range self {
		for btn, order := range floorOrders {
			minElement := order

			for _, peer := range peerList {
				if s.Orders[peer][floor][btn].AlterId > minElement.AlterId {
					minElement = s.Orders[peer][floor][btn]
				}
			}
			toReturn[floor][btn] = minElement
		}
	}
	return toReturn
}

package state

import "Driver-go/elevator-system/elevio"

// Cab calls
type CabCallStruct struct {
	Floor   int
	Active  bool
	AlterId uint8
}

// Hall Calls
type hallCallStruct struct {
	Floor   int
	Dir     int
	Active  bool
	AlterId uint8
}

// Struct for the Worldview
type StateStruct struct {
	Id        string                         //id of the elevator sending
	Elevators map[string]ElevatorStateStruct // Map of all seen elevators
}

// Struct containing the calls of each elevator
type ElevatorStateStruct struct {
	CabCalls  []CabCallStruct
	HallCalls []hallCallStruct
}

// "Constructor" for ElevatorState with all orders inactive.
func createElevatorState(numFloors int) ElevatorStateStruct {

	cabCalls := make([]CabCallStruct, numFloors)
	for i := 0; i < numFloors; i++ {
		cabCalls[i] = CabCallStruct{Floor: i, Active: false, AlterId: 0}
	}

	hallCalls := make([]hallCallStruct, numFloors*2)
	for i := 0; i < numFloors; i++ {
		hallCalls[2*i] = hallCallStruct{Floor: i, Dir: 0, Active: false, AlterId: 0}
		hallCalls[2*i+1] = hallCallStruct{Floor: i, Dir: 1, Active: false, AlterId: 0}
	}

	return ElevatorStateStruct{
		CabCalls:  cabCalls,
		HallCalls: hallCalls,
	}

}

func CreateStateStruct(id string, numFloors int) StateStruct {
	elevators := make(map[string]ElevatorStateStruct)
	elevators[id] = createElevatorState(numFloors)

	return StateStruct{
		Id:        id,
		Elevators: elevators,
	}
}

// Compare single elevator state
func (own *ElevatorStateStruct) compareIncoming(incoming ElevatorStateStruct) {
	for i, ownCabCall := range own.CabCalls {
		if incoming.CabCalls[i].AlterId > ownCabCall.AlterId { // How to handle cyclic wraparound?
			own.CabCalls[i] = incoming.CabCalls[i]
		}
	}
	for i, ownHallCall := range own.HallCalls {
		if incoming.HallCalls[i].AlterId > ownHallCall.AlterId {
			own.HallCalls[i] = incoming.HallCalls[i]
		}
	}

}

// Looks at incoming state and updates state based on alterId.
func (own *StateStruct) CompareIncoming(incoming StateStruct) {

	//For each elevator incoming knows about, update old info and add potential not known about elevators
	for key, incoming_val := range incoming.Elevators {
		own_val, exists := own.Elevators[key]
		if exists {
			own_val.compareIncoming(incoming_val)
			own.Elevators[key] = own_val
		} else {
			//Add it to your state without comparing
			own.Elevators[key] = incoming_val
		}
	}

	//Also, update hall calls so that new hall calls from incoming are propagated to ourselves
	ownHallCalls := own.Elevators[own.Id].HallCalls
	incomingHallCalls := incoming.Elevators[incoming.Id].HallCalls

	for i, ownVal := range ownHallCalls {
		if incomingHallCalls[i].AlterId > ownVal.AlterId { //How to handle cyclic wraparound?
			own.Elevators[own.Id].HallCalls[i] = incomingHallCalls[i]
		}
	}
}

// Sets an order in the elevator state
func (elev *ElevatorStateStruct) SetOrder(button elevio.ButtonEvent) {
	switch button.Button {
	case elevio.BT_HallUp:
		if !elev.HallCalls[2*button.Floor].Active {
			elev.HallCalls[2*button.Floor].Active = true
			elev.HallCalls[2*button.Floor].AlterId++
		}
	case elevio.BT_HallDown:
		if !elev.HallCalls[2*button.Floor+1].Active {
			elev.HallCalls[2*button.Floor+1].Active = true
			elev.HallCalls[2*button.Floor+1].AlterId++
		}
	case elevio.BT_Cab:
		if !elev.CabCalls[button.Floor].Active {
			elev.CabCalls[button.Floor].Active = true
			elev.CabCalls[button.Floor].AlterId++

		}
	}
}

// Sets an order at itself in the worldview state.
func (s *StateStruct) SetOrder(button elevio.ButtonEvent) {
	val, exists := s.Elevators[s.Id]
	if exists {
		val.SetOrder(button)
		s.Elevators[s.Id] = val
	} else {
		panic("Elevator state does not know about itself!")
	}
}

// Sets an order in the elevator state
func (elev *ElevatorStateStruct) UnsetOrder(button elevio.ButtonEvent) {
	switch button.Button {
	case elevio.BT_HallUp:
		if elev.HallCalls[2*button.Floor].Active {
			elev.HallCalls[2*button.Floor].Active = false
			elev.HallCalls[2*button.Floor].AlterId++
		}
	case elevio.BT_HallDown:
		if elev.HallCalls[2*button.Floor+1].Active {
			elev.HallCalls[2*button.Floor+1].Active = false
			elev.HallCalls[2*button.Floor+1].AlterId++
		}
	case elevio.BT_Cab:
		if elev.CabCalls[button.Floor].Active {
			elev.CabCalls[button.Floor].Active = false
			elev.CabCalls[button.Floor].AlterId++

		}
	}
}

// Unsets an order at itself in the worldview state
func (s *StateStruct) UnsetOrder(button elevio.ButtonEvent) {
	val, exists := s.Elevators[s.Id]
	if exists {
		val.SetOrder(button)
		s.Elevators[s.Id] = val
	} else {
		panic("Elevator state does not know about itself!")
	}
}

// Check if all peers know about a hall call order, and sends a message on a
// channel if this is the case.
//
// Not 100% sure this is the right way to do it, some issues might occur
func (s *StateStruct) SendNewHallOrders(peerList []string,
	newOrderCh chan<- elevio.ButtonEvent, completedOrderCh chan<- elevio.ButtonEvent) {
	for i, selfVal := range s.Elevators[s.Id].HallCalls {
		allEqual := true
		for _, peer := range peerList {
			if s.Elevators[peer].HallCalls[i].Active != selfVal.Active {
				allEqual = false
				break
			}
		}
		if allEqual {
			button := elevio.ButtonType(i % 2)
			if selfVal.Active {
				newOrderCh <- elevio.ButtonEvent{Floor: selfVal.Floor, Button: button}
			} else {
				completedOrderCh <- elevio.ButtonEvent{Floor: selfVal.Floor, Button: button}
			}
		}
	}
}

// Checks if some cab orders are known by everyone, and sends a message if this is the case.
func (s *StateStruct) SendNewCabOrders(peerList []string, newOrderCh chan<- elevio.ButtonEvent) {
	for i, selfVal := range s.Elevators[s.Id].CabCalls {
		if !selfVal.Active {
			continue
		}
		allEqual := true
		for _, peer := range peerList {
			if s.Elevators[peer].CabCalls[i].Active != selfVal.Active {
				allEqual = false
				break
			}
		}
		if allEqual {
			newOrderCh <- elevio.ButtonEvent{Floor: selfVal.Floor, Button: elevio.BT_Cab}
		}
	}
}

package assigner

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
)

// Channels:
// worldviewCh: gets latest state from all elevators
// ordersCh: sends an assigned order to all the elevators

// Gets all elevator states and hallassignments and sends updated orders to elevators
func RunAssigner(worldviewCh <-chan state.StateStruct, ordersCh chan<- state.ElevatorOrders, numFloors int) {
	for {
		//newest elevator state
		worldview := <-worldviewCh

		//Sets button lights on ordered floors
		for _, FloorOrders := range worldview.Orders[worldview.Id] {
			for _, order := range FloorOrders {
				elevio.SetButtonLamp(order.Order.Button, order.Order.Floor, order.Active)
			}
		}
		//assign new orders to elevators (if any)
		requests := AssignHallRequests(worldview, numFloors)
		if requests != nil {
			ordersCh <- requests
		} else {
		}
	}
}

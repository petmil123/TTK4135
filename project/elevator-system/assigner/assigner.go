// Gets state from the active peers and assigns orders to itself based on these states.
//
// Channels:
//
// worldviewCh: gets latest state from all elevators
//
// ordersCh: sends assigned orders to itself
package assigner

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
)

// Gets all elevator states and hall assignments and sends updated orders to elevator
func RunAssigner(worldviewCh <-chan state.StateStruct, ordersCh chan<- state.ElevatorOrders, numFloors int) {
	oldOrders := state.CreateElevatorOrders(numFloors) // The only reason for storing this is to not set all lights every time.
	for _, FloorOrders := range oldOrders {
		for _, order := range FloorOrders {
			elevio.SetButtonLamp(order.Order.Button, order.Order.Floor, order.Active)
		}
	}

	for {
		//newest worldview
		newWorldview := <-worldviewCh

		//Sets button lights on orders with new active status
		for i, FloorOrders := range newWorldview.Orders[newWorldview.Id] {
			for j, order := range FloorOrders {
				if oldOrders[i][j].Active != order.Active {
					elevio.SetButtonLamp(order.Order.Button, order.Order.Floor, order.Active)
				}
				oldOrders[i][j] = order // Deep copy on the fly
			}
		}
		//assign new orders to elevator (if any)
		requests := AssignHallRequests(newWorldview, numFloors)
		if requests != nil {
			ordersCh <- requests
		}

	}
}

package assigner

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
)

func RunAssigner(worldviewCh <-chan state.StateStruct, ordersCh chan<- state.ElevatorOrders, numFloors int) {
	for {
		worldview := <-worldviewCh

		for _, FloorOrders := range worldview.Orders[worldview.Id] {
			for _, order := range FloorOrders {
				elevio.SetButtonLamp(order.Order.Button, order.Order.Floor, order.Active)
			}
		}
		requests := AssignHallRequests(worldview, numFloors)
		if requests != nil {
			ordersCh <- requests
		} else {
		}
	}
}

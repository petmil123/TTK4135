package assigner

import (
	"Driver-go/elevator-system/elevio"
	"Driver-go/elevator-system/state"
)

func RunAssigner(worldviewCh <-chan state.StateStruct, ordersCh chan<- state.ElevatorOrders) {
	for {

		worldview := <-worldviewCh
		ordersCh <- AssignHallRequests(worldview)
		for _, FloorOrders := range worldview.Orders[worldview.Id] {
			for _, order := range FloorOrders {
				elevio.SetButtonLamp(order.Order.Button, order.Order.Floor, order.Active)
			}
		}
	}
}

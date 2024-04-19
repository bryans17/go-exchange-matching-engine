package main

func ordToInput(order Order) input {
	return input{order.order_type, order.order_id, order.price, order.count, order.instrument}
}

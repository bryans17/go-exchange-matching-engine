package main

import "container/heap"

type Order struct {
	order_type   inputType
	order_id     uint32
	price        uint32
	count        uint32
	execution_id uint32
	time_added   int64
	instrument   string
}

// To build a priority queue, implement the Heap interface with the (negative) priority as the ordering for the Less method,
// so Push adds items while Pop removes the highest-priority item from the queue.

type OrderBook struct {
	book []*Order
}

func (orderbook *OrderBook) Top() Order {
	return *orderbook.book[0]
}

func (orderbook OrderBook) Len() int { return len(orderbook.book) }

func (orderbook OrderBook) Less(i, j int) bool {
	if orderbook.book[i].price == orderbook.book[j].price {
		return orderbook.book[i].time_added < orderbook.book[j].time_added
	} else if orderbook.book[i].order_type == inputSell {
		return orderbook.book[i].price < orderbook.book[j].price
	} else {
		return orderbook.book[i].price > orderbook.book[j].price
	}
}

func (orderbook OrderBook) Swap(i, j int) {
	orderbook.book[i], orderbook.book[j] = orderbook.book[j], orderbook.book[i]
}

func (orderbook *OrderBook) Push(x any) {
	item := x.(*Order)
	orderbook.book = append(orderbook.book, item)
}

func (orderbook *OrderBook) Pop() any {
	old := orderbook.book
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	orderbook.book = old[0 : n-1]
	return item
}

func (orderbook *OrderBook) Remove(order *Order) bool {
	for i := 0; i < orderbook.Len(); i++ {
		if orderbook.book[i].order_id == order.order_id {
			heap.Remove(orderbook, i)
			return true
		}
	}
	return false
}

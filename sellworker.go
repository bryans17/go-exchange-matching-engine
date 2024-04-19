package main

import (
	"container/heap"
)

type SellWorker struct {
	in  <-chan Order
	add <-chan Order
	// the toAdd channel of the SellWorker is the add channel of the BuyWorker
	toAdd        chan<- Order
	buyout       chan<- struct{}
	buyOrderBook OrderBook
}

func (sellWorker *SellWorker) start() {
	for {
		select {
		case ord := <-sellWorker.in:
			switch ord.order_type {
			case inputSell:
				sellWorker.matchSell(&ord)
			case inputCancel:
				sellWorker.cancelOrder(&ord)
			}
			sellWorker.toAdd <- ord
		case toInsert := <-sellWorker.add:
			if toInsert.count > 0 {
				sellWorker.insertBuy(toInsert)
			}
			sellWorker.buyout <- struct{}{}
		}
	}
}

func (sellWorker *SellWorker) cancelOrder(toCancel *Order) {
	isRemoved := sellWorker.buyOrderBook.Remove(toCancel)
	outputOrderDeleted(ordToInput(*toCancel), isRemoved, toCancel.time_added)
	toCancel.count = 0
}

func (sellWorker *SellWorker) insertBuy(toInsert Order) {
	heap.Push(&sellWorker.buyOrderBook, &toInsert)
	outputOrderAdded(ordToInput(toInsert), toInsert.time_added)
}

func (sellWorker *SellWorker) matchSell(toMatch *Order) {
	for sellWorker.buyOrderBook.Len() > 0 {
		if toMatch.count == 0 {
			break
		}

		topOfBuyBook := sellWorker.buyOrderBook.Top()
		if topOfBuyBook.price < toMatch.price {
			break
		}
		heap.Pop(&sellWorker.buyOrderBook)

		matchedQty := min(toMatch.count, topOfBuyBook.count)
		toMatch.count -= matchedQty
		topOfBuyBook.count -= matchedQty

		if matchedQty > 0 {
			outputOrderExecuted(
				topOfBuyBook.order_id,
				toMatch.order_id,
				topOfBuyBook.execution_id,
				topOfBuyBook.price,
				matchedQty,
				toMatch.time_added,
			)
		}
		if topOfBuyBook.count > 0 {
			topOfBuyBook.execution_id++
			heap.Push(&sellWorker.buyOrderBook, &topOfBuyBook)
		}
	}
}

package main

import (
	"container/heap"
)

type BuyWorker struct {
	in  <-chan Order
	add <-chan Order
	// the toAdd channel of the BuyWorker is the add channel of the SellWorker
	toAdd         chan<- Order
	sellout       chan<- struct{}
	sellOrderBook OrderBook
}

func (buyWorker *BuyWorker) start() {
	for {
		select {
		case ord := <-buyWorker.in:
			switch ord.order_type {
			case inputBuy:
				buyWorker.matchBuy(&ord)
			case inputCancel:
				buyWorker.cancelOrder(&ord)
			}
			buyWorker.toAdd <- ord
		case toInsert := <-buyWorker.add:
			if toInsert.count > 0 {
				buyWorker.insertSell(toInsert)
			}
			buyWorker.sellout <- struct{}{}
		}
	}
}

func (buyWorker *BuyWorker) cancelOrder(toCancel *Order) {
	isRemoved := buyWorker.sellOrderBook.Remove(toCancel)
	outputOrderDeleted(ordToInput(*toCancel), isRemoved, toCancel.time_added)
	toCancel.count = 0
}

func (buyWorker *BuyWorker) insertSell(toInsert Order) {
	heap.Push(&buyWorker.sellOrderBook, &toInsert)
	outputOrderAdded(ordToInput(toInsert), toInsert.time_added)
}

func (buyWorker *BuyWorker) matchBuy(toMatch *Order) {
	for buyWorker.sellOrderBook.Len() > 0 {
		if toMatch.count == 0 {
			break
		}

		topOfSellBook := buyWorker.sellOrderBook.Top()
		if topOfSellBook.price > toMatch.price {
			break
		}
		heap.Pop(&buyWorker.sellOrderBook)

		matchedQty := min(toMatch.count, topOfSellBook.count)
		toMatch.count -= matchedQty
		topOfSellBook.count -= matchedQty

		if matchedQty > 0 {
			outputOrderExecuted(
				topOfSellBook.order_id,
				toMatch.order_id,
				topOfSellBook.execution_id,
				topOfSellBook.price,
				matchedQty,
				toMatch.time_added,
			)
		}
		if topOfSellBook.count > 0 {
			topOfSellBook.execution_id++
			heap.Push(&buyWorker.sellOrderBook, &topOfSellBook)
		}
	}
}

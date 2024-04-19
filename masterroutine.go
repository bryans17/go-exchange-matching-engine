package main

type MasterRoutine struct {
	incomingOrders chan Order
	sellCount      uint32
	buyCount       uint32
	sellin         chan Order
	buyin          chan Order
	sellout        chan struct{}
	buyout         chan struct{}
	typeByOrderId  map[uint32]inputType
}

func (master *MasterRoutine) initBuyAndSellRoutines() {
	// init buy and sell add channels
	var selladd chan Order = make(chan Order, 10000)
	var buyadd chan Order = make(chan Order, 10000)

	var sellWorker *SellWorker = &SellWorker{
		in:     master.sellin,
		add:    selladd,
		toAdd:  buyadd,
		buyout: master.buyout,
		buyOrderBook: OrderBook{
			book: make([]*Order, 0),
		},
	}
	var buyWorker *BuyWorker = &BuyWorker{
		in:      master.buyin,
		add:     buyadd,
		toAdd:   selladd,
		sellout: master.sellout,
		sellOrderBook: OrderBook{
			book: make([]*Order, 0),
		},
	}

	//start running the buy and sell routines
	go sellWorker.start()
	go buyWorker.start()
}

func (master *MasterRoutine) start() {
	master.initBuyAndSellRoutines()
	for {
		select {
		case ord := <-master.incomingOrders:
			if ord.order_type != inputCancel {
				master.typeByOrderId[ord.order_id] = ord.order_type
			}
			switch ord.order_type {
			case inputBuy:
				master.handleBuy(ord)
			case inputSell:
				master.handleSell(ord)
			case inputCancel:
				origType, ok := master.typeByOrderId[ord.order_id]
				if !ok {
					outputOrderDeleted(ordToInput(ord), false, ord.time_added)
				} else {
					switch origType {
					case inputBuy:
						master.handleSell(ord)
					case inputSell:
						master.handleBuy(ord)
					}
				}
			}
		case <-master.sellout:
			master.sellCount--
		case <-master.buyout:
			master.buyCount--
		}
	}
}

func (master *MasterRoutine) handleSell(ord Order) {
	for master.buyCount > 0 {
		<-master.buyout
		// decrement buycount
		master.buyCount--
	}
	master.sellCount++
	master.sellin <- ord
}

func (master *MasterRoutine) handleBuy(ord Order) {
	for master.sellCount > 0 {
		<-master.sellout
		// decrement buycount
		master.sellCount--
	}
	master.buyCount++
	master.buyin <- ord
}

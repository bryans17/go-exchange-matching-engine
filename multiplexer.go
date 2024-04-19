package main

type Multiplexer struct {
	incomingOrders     chan Order
	instrumentToMaster map[string]*MasterRoutine
	instrByOrderId     map[uint32]string
}

func (mux *Multiplexer) start() {
	var time int64 = 0
	for {
		ord := <-mux.incomingOrders
		time++
		ord.time_added = time

		if ord.order_type != inputCancel {
			mux.instrByOrderId[ord.order_id] = ord.instrument
		} else {
			instr, ok := mux.instrByOrderId[ord.order_id]
			if !ok {
				outputOrderDeleted(ordToInput(ord), false, ord.time_added)
			}
			ord.instrument = instr
		}
		if mas, ok := mux.instrumentToMaster[ord.instrument]; ok {
			mas.incomingOrders <- ord
		} else {
			// create a new orderBook master go routine for this instrument
			mas := mux.initNewMasterRoutine()
			mux.instrumentToMaster[ord.instrument] = mas

			go mas.start()

			mas.incomingOrders <- ord
		}
	}
}

func (mux *Multiplexer) initNewMasterRoutine() *MasterRoutine {
	inputChannelCapacity, outputChannelCapacity := 10000, 10000

	masterRoutine := MasterRoutine{
		incomingOrders: make(chan Order, inputChannelCapacity),
		sellin:         make(chan Order),
		buyin:          make(chan Order),
		sellout:        make(chan struct{}, outputChannelCapacity),
		buyout:         make(chan struct{}, outputChannelCapacity),
		typeByOrderId:  make(map[uint32]inputType),
	}

	return &masterRoutine
}

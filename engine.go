package main

import "C"
import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Engine struct {
	mux Multiplexer
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go e.handleConn(conn)
}

func (e *Engine) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		in, err := readInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		// order -> multiplexer -> masterroutine for that instrument -> workerroutine for that order type
		order := Order{
			order_type:   in.orderType,
			price:        in.price,
			count:        in.count,
			order_id:     in.orderId,
			instrument:   in.instrument,
			execution_id: 1,
		}
		e.mux.incomingOrders <- order
	}
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

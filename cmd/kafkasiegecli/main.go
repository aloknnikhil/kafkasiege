package main

import (
	"flag"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var (
	brokerEP string
	conns    uint
)

func init() {
	flag.StringVar(&brokerEP, "broker-endpoint", "", "Broker Endpoint (including port)")
	flag.UintVar(&conns, "connections", 1, "No. of connections to launch a siege with")
}

func main() {
	flag.Parse()
	if len(brokerEP) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	completed := uint64(0)

	for i := uint(0); i < conns; i++ {
		go func() {
			defer func() {
				atomic.AddUint64(&completed, 1)
			}()
			conn, err := net.Dial("tcp", brokerEP)
			if err != nil {
				log.Printf("TCP dial error: %s\n", err.Error())
				return
			}
			if err = conn.Close(); err != nil {
				log.Printf("Connection close error: %s\n", err.Error())
			}
		}()
	}

	for atomic.LoadUint64(&completed) < uint64(conns) {
		log.Printf("Waiting on %d threads to complete\n", uint64(conns)-atomic.LoadUint64(&completed))
		time.Sleep(100 * time.Millisecond)
	}
}

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
	timeout  uint
	retries  uint
)

func init() {
	flag.StringVar(&brokerEP, "broker-endpoint", "", "Broker Endpoint (including port)")
	flag.UintVar(&conns, "connections", 1, "No. of connections to launch a siege with")
	flag.UintVar(&timeout, "connect-timeout", 1, "Connection timeout (in ms)")
	flag.UintVar(&retries, "retries", 1, "No. of connect retries")
}

func main() {
	flag.Parse()
	if len(brokerEP) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	connected := uint64(0)
	failed := uint64(0)

	for i := uint(0); i < conns; i++ {
		go func() {
			retryCount := uint(0)
			var conn net.Conn
			var err error
			defer func() {
				if err != nil {
					atomic.AddUint64(&failed, 1)
				} else {
					atomic.AddUint64(&connected, 1)
				}
			}()

			for retryCount <= retries {
				conn, err = net.DialTimeout("tcp", brokerEP, time.Millisecond*time.Duration(timeout))
				if err != nil {
					log.Printf("TCP dial error: %s\n", err.Error())
					retries++
				}
			}
			if retryCount == retries && err != nil {
				log.Println("Exhausted retries")
				return
			}
			if err = conn.Close(); err != nil {
				log.Printf("Connection close error: %s\n", err.Error())
				return
			}
		}()
	}

	total := uint64(0)
	for total < uint64(conns) {
		log.Printf("Waiting on %d threads to complete\n", uint64(conns)-total)
		time.Sleep(100 * time.Millisecond)
		total = atomic.LoadUint64(&connected) + atomic.LoadUint64(&failed)
	}

	log.Printf("Total: %d \t Success: %d \t Error: %d\n", conns, connected, failed)
}

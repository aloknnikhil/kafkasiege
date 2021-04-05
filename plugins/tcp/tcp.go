package main

import (
	"github.com/aloknnikhil/kafkasiege/pkg/harness"
	"github.com/pelletier/go-toml"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	pluginName      = "tcp"
	failedMetric    = "failed"
	connectedMetric = "connected"
)

// Plugin - Exported reference to harness.Plugin implementation
var Plugin TCP

func main() {
	panic("This is not an executable. Build it as a plugin w/ '-buildmode=plugin'")
}

//type Config struct {
//}

type TCP struct {
	harnessImpl harness.Harness
}

func (t *TCP) Init(harnessImpl harness.Harness) (scheduler harness.Scheduler, err error) {
	t.harnessImpl = harnessImpl

	// Load plugin-specific config
	// TODO: Skip for now

	// Set scheduler
	scheduler = harness.Default
	return
}

func (t *TCP) Name() (name string) {
	return pluginName
}

// TODO: Parse tree
func (t *TCP) Config() (config *toml.Tree) {
	return nil
}

func (t *TCP) Function() harness.Func {
	return func(connectionId uint64) {
		retryCount := uint(0)
		var conn net.Conn
		var err error
		defer func() {
			if t.harnessImpl == nil {
				panic("harness is nil")
			}

			if t.harnessImpl.Metrics() == nil {
				panic("metrics is nil")
			}
			if err != nil {
				t.harnessImpl.Metrics().Count(failedMetric, 1)
			} else {
				t.harnessImpl.Metrics().Count(connectedMetric, 1)
			}
		}()

		for retryCount <= t.harnessImpl.Config().Retries {
			if conn, err = net.DialTimeout("tcp", t.harnessImpl.Config().BrokerEndpoint,
				time.Duration(t.harnessImpl.Config().Timeout)*time.Millisecond); err != nil {
				log.Printf("[ID: %d] TCP dial error: %s\n", connectionId, err.Error())
				retryCount++
				if retryCount > t.harnessImpl.Config().Retries {
					break
				}
				retryIn := rand.Intn(1000)
				log.Printf("[ID: %d] Will retry in %d ms\n", connectionId, retryIn)
				time.Sleep(time.Duration(retryIn) * time.Millisecond)
			} else {
				break
			}
		}
		if retryCount > t.harnessImpl.Config().Retries && err != nil {
			log.Printf("[ID: %d] Exhausted retries\n", connectionId)
			return
		}
		if err = conn.Close(); err != nil {
			log.Printf("Connection close error: %s\n", err.Error())
			return
		}
	}
}

func (t *TCP) Run() {
	panic("not implemented")
}

func (t *TCP) Stop() (err error) {
	panic("not implemented")
}

func (t *TCP) Done() bool {
	connected := t.harnessImpl.Metrics().Get(connectedMetric)
	failed := t.harnessImpl.Metrics().Get(failedMetric)
	remaining := t.harnessImpl.Config().Connections - uint64(connected+failed)
	log.Printf("Waiting on %d connections to complete\n", remaining)
	return remaining == 0
}

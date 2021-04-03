package harness

import (
	"bytes"
	"io"
	"math/rand"
	"sync"
	"time"
)

const (
	payloadSize = 20 * 1024 * 1024 // Bytes
)

var payloadData []byte

type Generator interface {
	Payload(api API) io.Reader
}

type BinaryPayloadGenerator struct {
	sync.Once
}

func (m *BinaryPayloadGenerator) Payload(api API) (payload io.Reader) {
	switch api {
	case Binary:
		m.Do(func() {
			if payloadData == nil {
				payloadData = make([]byte, payloadSize)
				rand.Seed(time.Now().UnixNano())
				rand.Read(payloadData)
			}
		})
		payload = bytes.NewReader(payloadData)
	default:
		panic("not implemented")
	}
	return
}

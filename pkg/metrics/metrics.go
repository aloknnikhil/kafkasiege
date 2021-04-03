package metrics

import (
	"sync"
	"sync/atomic"
)

type Metrics interface {
	Count(metricName string, value int64)
	Get(metricName string) int64
	Snapshot() map[string]int64
}

// Thread-safe metrics implementation
type Basic struct {
	Metrics
	records sync.Map
}

func (b *Basic) Count(metricName string, value int64) {
	val, loaded := b.records.LoadOrStore(metricName, &value)
	if loaded {
		atomic.AddInt64(val.(*int64), value)
	}
}

func (b *Basic) Get(metricName string) (value int64) {
	val, loaded := b.records.Load(metricName)
	if loaded {
		value = *val.(*int64)
	} else {
		value = 0
	}
	return
}

func (b *Basic) Snapshot() (snapshot map[string]int64) {
	snapshot = make(map[string]int64)

	b.records.Range(func(key, value interface{}) bool {
		snapshot[key.(string)] = *(value.(*int64))
		return true
	})
	return
}

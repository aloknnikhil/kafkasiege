package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aloknnikhil/kafkasiege/pkg/harness"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

const (
	pluginName               = "kafka"
	failedMetric             = "failed"
	producedMetric           = "produced"
	defaultTopic             = "topic"
	defaultReplicationFactor = 3
	defaultPartitions        = 100
)

// Plugin - Exported reference to harness.Plugin implementation
var Plugin Kafka

var securityConfig kafka.ConfigMap = kafka.ConfigMap{
	// "security.protocol": "SSL",
	// "ssl.ca.location":   "/opt/benchmark/ca.pem",
	//"sasl.mechanism":    "PLAIN",
	//"sasl.jaas.config": `org.apache.kafka.common.security.plain.PlainLoginModule required
	//username="client"
	//password="client-secret";`,
}

func main() {
	panic("This is not an executable. Build it as a plugin w/ '-buildmode=plugin'")
}

//type Config struct {
//}

type Kafka struct {
	harnessImpl harness.Harness
	config      kafka.ConfigMap
}

func (k *Kafka) Init(harnessImpl harness.Harness) (scheduler harness.Scheduler, err error) {
	k.harnessImpl = harnessImpl
	var adminClient *kafka.AdminClient
	securityConfig.SetKey("bootstrap.servers", harnessImpl.Config().BrokerEndpoint)

	if adminClient, err = kafka.NewAdminClient(&securityConfig); err != nil {
		err = errors.Wrap(err, "kafka.NewAdminClient()")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		err = errors.Wrap(err, "time.ParseDuration(60s)")
		return
	}

	// Attempt to delete the topic if it already exists
	_, err = adminClient.DeleteTopics(
		ctx, []string{defaultTopic}, kafka.SetAdminOperationTimeout(maxDur))
	log.Printf("[WARN] - Failed to delete topic: %s\n", defaultTopic)

	creates, err := adminClient.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             defaultTopic,
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplicationFactor}},
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		err = errors.Wrapf(err, "[FATAL] Failed to create topic: %s\n", defaultTopic)
		return
	}

	for _, create := range creates {
		fmt.Printf("Created: %s\n", create)
	}

	return
}

func (k *Kafka) Name() string {
	return pluginName
}

func (k *Kafka) Config() *toml.Tree {
	return nil
}

func (k *Kafka) Function() harness.Func {
	return func(connectionId uint64) {
		var err error
		var producer *kafka.Producer
		defer func() {
			if err != nil {
				log.Printf("[Connection: %d] [WARN] %s", connectionId, err.Error())
				k.harnessImpl.Metrics().Count(failedMetric, 1)
			} else {
				k.harnessImpl.Metrics().Count(producedMetric, 1)
			}

			if producer != nil {
				producer.Close()
			}
		}()
		if producer, err = kafka.NewProducer(&securityConfig); err != nil {
			err = errors.Wrap(err, "kafka.NewProducer()")
			return
		}

		go func() {
			for e := range producer.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					dt := time.Now()
					if ev.TopicPartition.Error != nil {
						log.Printf("[%+v][Connection: %d] Delivery failed: %+v\n", dt.Format(time.UnixDate), connectionId, ev.TopicPartition)
					} else {
						log.Printf("[%+v][Connection: %d] Delivered message", dt.Format(time.UnixDate), connectionId)
					}
				}
			}
		}()

		sendToTopic := defaultTopic
		for i := 0; ; i++ {
			time.Sleep(1 * time.Second)
			if err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &sendToTopic,
					Partition: 1,
				},
				Value: []byte(fmt.Sprintf("%d", i)),
			}, nil); err != nil {
				err = errors.Wrap(err, "prodcuer.Produce()")
				return
			}

			// Flush every message
			if i%1 == 0 {
				producer.Flush(15 * 1000)
			}
		}
	}
}

func (k *Kafka) Run() {
	panic("not implemented")
}

func (k *Kafka) Stop() error {
	return nil
}

func (k *Kafka) Done() bool {
	connected := k.harnessImpl.Metrics().Get(producedMetric)
	failed := k.harnessImpl.Metrics().Get(failedMetric)
	remaining := k.harnessImpl.Config().Connections - uint64(connected+failed)
	// log.Printf("Waiting on %d producers to complete\n", remaining)
	return remaining == 0
}

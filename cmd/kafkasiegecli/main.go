package main

import (
	"flag"
	"github.com/aloknnikhil/kafkasiege/pkg/harness"
	"github.com/pkg/errors"
	"log"
	"os"
)

var (
	configPath  string
	testHarness harness.Harness
)

func init() {
	flag.StringVar(&configPath, "config", "", "Path to the TOML config")
}

func main() {
	flag.Parse()
	if len(configPath) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	var err error
	testHarness = &harness.Standard{}
	if err = testHarness.Init(configPath); err != nil {
		panic(err)
	}

	if err = testHarness.Run(); err != nil {
		err = errors.Wrap(err, "testHarness.Run()")
		panic(err)
	}

	metrics := testHarness.Metrics().Snapshot()
	for metric, value := range metrics {
		log.Printf("Metric: %s\t Value: %d\n", metric, value)
	}
}

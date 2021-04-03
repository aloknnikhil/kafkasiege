package harness

import (
	"fmt"
	"github.com/aloknnikhil/kafkasiege/pkg/config"
	"github.com/aloknnikhil/kafkasiege/pkg/metrics"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	goplugin "plugin"
	"reflect"
	"time"
)

type Scheduler int

const (
	Default Scheduler = iota
	Custom
)

type Harness interface {
	Generator() Generator
	Config() *config.Config
	Init(configPath string) error
	LoadPlugin(libraryPath string) (Plugin, error)
	Metrics() metrics.Metrics
	Run() error
	Stop()
}

type Standard struct {
	plugins   []Plugin
	config    *config.Config
	generator Generator
	metrics   metrics.Metrics
}

func (s *Standard) Generator() Generator {
	return s.generator
}

func (s *Standard) Config() *config.Config {
	return s.config
}

func (s *Standard) Init(configPath string) (err error) {
	// Parse config
	var data []byte
	if data, err = ioutil.ReadFile(configPath); err != nil {
		err = errors.Wrapf(err, "ioutil.ReadFile(%s)", configPath)
		return
	}
	s.config = &config.Config{}
	if err = toml.Unmarshal(data, s.config); err != nil {
		err = errors.Wrapf(err, "toml.Unmarshal(%s)", configPath)
		return
	}
	if err = s.config.Validate(); err != nil {
		err = errors.Wrap(err, "config.Validate()")
		return
	}

	// Load plugins
	for _, pluginCfg := range s.config.Plugin {
		var pluginImpl Plugin
		if pluginImpl, err = s.LoadPlugin(pluginCfg.Path); err != nil {
			err = errors.Wrapf(err, "s.LoadPlugin(%s)", pluginCfg.Name)
			return
		}
		s.plugins = append(s.plugins, pluginImpl)
	}

	// Setup harness
	s.metrics = &metrics.Basic{}
	s.generator = &BinaryPayloadGenerator{}
	return
}

func (s *Standard) LoadPlugin(libraryPath string) (pluginImpl Plugin, err error) {
	var pluginLib *goplugin.Plugin
	if pluginLib, err = goplugin.Open(libraryPath); err != nil {
		err = errors.Wrapf(err, "plugin.Open(%s)", libraryPath)
		return
	}

	// Lookup exported plugin.Plugin instance
	var val goplugin.Symbol
	if val, err = pluginLib.Lookup(ExportedPluginReference); err != nil {
		err = errors.Wrapf(err, "pluginLib.Lookup(%s)", ExportedPluginReference)
		return
	}

	var ok bool
	if pluginImpl, ok = val.(Plugin); !ok {
		err = errors.Errorf("Exported plugin var not of type harness.Plugin was instead: %s",
			reflect.TypeOf(val).Name())
		return
	}
	return
}

func (s *Standard) Metrics() metrics.Metrics {
	return s.metrics
}

func (s *Standard) Run() (err error) {
	fmt.Println("Each pluginImpl will be executed sequentially")
	for _, pluginImpl := range s.plugins {
		var scheduler Scheduler
		if scheduler, err = pluginImpl.Init(s); err != nil {
			err = errors.Wrapf(err, "pluginImpl.Init(%s)", pluginImpl.Name())
			return
		}

		switch scheduler {
		case Default:
			s.schedule(pluginImpl.Function())
			for !pluginImpl.Done() {
				log.Printf("Waiting for %s plugin to finish\n", pluginImpl.Name())
				time.Sleep(100 * time.Millisecond)
			}
		case Custom:
			pluginImpl.Run()
		default:
			err = errors.Errorf("Unknown scheduler type: %d", scheduler)
			return
		}
	}
	return
}

func (s *Standard) Stop() {
	panic("not implemented")
}

func (s *Standard) schedule(p Func) {
	for i := uint64(0); i < s.config.Connections; i++ {
		go p(i)
	}
}

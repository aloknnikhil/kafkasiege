package config

import (
	"fmt"
	"github.com/pkg/errors"
)

type Plugin struct {
	Name   string
	Path   string
	Config string
}

func (p Plugin) Validate() (err error) {
	if len(p.Name) == 0 {
		err = errors.Errorf("Plugin name missing")
		return
	}
	if len(p.Path) == 0 {
		err = errors.Errorf("Plugin path missing")
		return
	}
	return
}

type Config struct {
	BrokerEndpoint string
	Connections    uint64
	Timeout        uint
	Retries        uint
	Plugin         []Plugin
}

func (c *Config) Validate() (err error) {
	if len(c.BrokerEndpoint) == 0 {
		err = errors.Errorf("Invalid broker endpoint")
		return
	}
	for _, config := range c.Plugin {
		if err = config.Validate(); err != nil {
			err = errors.Wrap(err, "plugin.Validate()")
			return
		}
	}
	return
}

func (c *Config) String() (out string) {
	out = fmt.Sprintf("Broker End-point: %s\t"+
		"Connections: %d\t Timeout: %d\t Retries: %d\t",
		c.BrokerEndpoint, c.Connections, c.Timeout, c.Retries)
	var pluginOut string
	for _, config := range c.Plugin {
		pluginOut = fmt.Sprintf("%s\nPlugin: %s\n Config: %+v", pluginOut, config.Name, config)
	}
	out = fmt.Sprintf("---------- Harness Config ---------- \n%s\n"+
		"---------- Plugin Config ---------- \n%s\n", out, pluginOut)
	return
}

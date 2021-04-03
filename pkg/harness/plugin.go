package harness

import (
	"github.com/pelletier/go-toml"
)

const (
	ExportedPluginReference = "Plugin"
)

type Func func(connectionId uint64)

type Plugin interface {
	Init(harness Harness) (Scheduler, error)
	Name() string
	Config() *toml.Tree

	// Function the default scheduler uses to invoke this plugin's workload
	// Called for every connection
	Function() Func

	// Called when using a custom scheduler
	// Called for every connection
	Run()
	Stop() error
	Done() bool
}

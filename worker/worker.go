package worker

import (
	"fmt"
	"runtime"
	"time"

	"github.com/labbcb/rnnr/db"

	"github.com/labbcb/rnnr/docker"
	"github.com/labbcb/rnnr/node"
	"github.com/labbcb/rnnr/server"
	"github.com/pbnjay/memory"
)

// Worker server is a standalone task executor that can be connected with a Master server.
type Worker struct {
	*server.Server
	Info *node.Info
}

// New creates a standalone worker server and initializes TES API endpoints.
// If cpuCores of ramGb is zero then the function will guess the maximum values.
func New(uri string, cpuCores int, ramGb float64) (*Worker, error) {
	client, err := db.Connect(uri, "rnnr-worker")
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB: %w", err)
	}

	rnnr, err := docker.Connect()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Docker: %w", err)
	}

	if cpuCores == 0 {
		cpuCores = runtime.NumCPU()
	}
	if ramGb == 0 {
		ramGb = float64(memory.TotalMemory() / 1e+9)
	}

	worker := &Worker{
		Server: server.New(client, rnnr),
		Info: &node.Info{
			CPUCores: cpuCores,
			RAMGb:    ramGb,
		},
	}

	worker.register()
	worker.Start(5 * time.Second)
	return worker, nil
}

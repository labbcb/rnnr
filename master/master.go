package master

import (
	"fmt"
	"github.com/labbcb/rnnr/db"
	"time"

	"github.com/labbcb/rnnr/server"
)

// Master is a master instance.
type Master struct {
	*server.Server
}

// New creates a server and initializes TES API and Node management endpoints.
func New(uri string) (*Master, error) {
	client, err := db.Connect(uri, "rnnr-master")
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	master := &Master{
		Server:       server.New(client, &Remote{}),
	}
	master.register()
	master.Start(5 * time.Second)
	return master, nil
}

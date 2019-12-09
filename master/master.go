package master

import (
	"errors"
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/db"
	"github.com/labbcb/rnnr/task"
	"log"
	"time"

	"github.com/labbcb/rnnr/server"
)

// Master is a master instance.
type Master struct {
	*server.Server
}

// New creates a server and initializes TES API and Node management endpoints.
// database is URI to MongoDB (without database name, which is 'rnnr-master')
func New(database string) (*Master, error) {
	connection, err := db.Connect(database, "rnnr-master")
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	master := &Master{
		Server: server.New(connection, &Remote{}),
	}
	master.register()
	go master.StartMonitor(5 * time.Second)
	return master, nil
}

func (m *Master) StartMonitor(sleepTime time.Duration) {
	for {
		if err := m.InitializeTasks(); err != nil {
			log.Println(err)
		}

		if err := m.RunTasks(); err != nil {
			log.Println(err)
		}

		if err := m.CheckTasks(); err != nil {
			log.Println(err)
		}

		time.Sleep(sleepTime)
	}
}

// InitializeTasks will iterate over queued tasks requesting a compute node to execute them.
// If node has free computing resources enough function will assign the node to this task and
// the task state will be set as initializing.
func (m *Master) InitializeTasks() error {
	queuedTasks, err := m.DB.FindByState(task.Queued)
	if err != nil {
		log.Println("could not get queued tasks:", err)
	}
	for _, t := range queuedTasks {
		n, err := m.Request(t.Resources)
		switch err.(type) {
		case nil:
			t.RemoteHost = n.Host
			t.State = task.Initializing
			if err := m.DB.Update(t); err != nil {
				log.Println("unable to update task:", err)
			}
			log.Println(t)
		case NoActiveNodes:
		case NoEnoughResources:
		default:
			return fmt.Errorf("could not request resources for task %s: %w", t.ID, err)
		}
	}

	return nil
}

// RunTask will check computing node and delegate task execution to the node.
// If node is not active or not responding them the task will be put as queued.
// Node will be disabled if not responding.
func (m *Master) RunTask(t *task.Task) {
	n, err := m.DB.GetByHost(t.RemoteHost)
	if err != nil {
		log.Println("unable to get node by host:", err)
		return
	}
	// Check has been disabled between task initialization and execution.
	if n.Active {
		if err := m.Runner.Run(t); err != nil {
			// If there are some network error then disable node.
			if _, ok := errors.Unwrap(err).(*client.NetworkError); ok {
				if err := m.Deactivate(n.ID); err != nil {
					log.Println(err)
				}
			}
			log.Println(err)
		}
	} else {
		t.State = task.Queued
	}
	// Update task state.
	if err := m.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	log.Println(t)
}

// RunTasks will iterate over initializing tasks and starting them concurrently.
func (m *Master) RunTasks() error {
	initializingTasks, err := m.DB.FindByState(task.Queued)
	if err != nil {
		return fmt.Errorf("could not get initializing tasks: %w", err)
	}

	for _, t := range initializingTasks {
		go m.RunTask(t)
	}

	return nil
}

// CheckTasks will iterate over running tasks checking if they have been completed well or not.
// It runs concurrently.
func (m *Master) CheckTasks() error {
	runningTasks, err := m.DB.FindByState(task.Running)
	if err != nil {
		return fmt.Errorf("getting running tasks: %w", err)
	}

	for _, t := range runningTasks {
		go m.CheckTask(t)
	}

	return nil
}

// CheckTask will check is a given tasks has been completed.
func (m *Master) CheckTask(t *task.Task) {
	n, err := m.DB.GetByHost(t.RemoteHost)
	if err != nil {
		log.Println("unable to get node by host:", err)
		return
	}
	if !n.Active {
		t.State = task.Queued
	} else {
		if err := m.Runner.Check(t); err != nil {
			if _, ok := errors.Unwrap(err).(*client.NetworkError); ok {
				if err := m.Deactivate(n.ID); err != nil {
					log.Println(err)
				}
			}
			log.Println(err)
		}
	}
	if err := m.DB.Update(t); err != nil {
		log.Println("unable to update task:", err)
	}
	if t.State != task.Running {
		log.Println(t)
	}
}

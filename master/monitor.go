package master

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/node"
	"github.com/labbcb/rnnr/task"
)

// Start starts two monitoring systems that will
// check state of the tasks communicating with database
// and task runner.
// System 1 will iterate over queued tasks initializing
// and running them. Second system will check running
// tasks to check they were not completed yet.
// Both systems will sleep when iteration is completed
// and start over again.
// Every action happens concurrently.
func (m *Master) Start(sleepTime time.Duration) {
	runQueuedTasks := func() {
		for {
			if err := m.initializeTasks(); err != nil {
				log.Println(err)
			}
			time.Sleep(sleepTime)
		}
	}
	checkRunningTasks := func() {
		for {
			ts, err := m.DB.FindByState(task.Running)
			if err != nil {
				log.Println("getting running tasks:", err)
				continue
			}
			for _, t := range ts {
				go m.check(t)
			}
			time.Sleep(sleepTime)
		}
	}
	updateUsage := func() {
		for {
			usage := make(map[string]*node.Usage)
			ts, err := m.DB.FindByState(task.Initializing, task.Running)
			if err != nil {
				log.Println("getting initializing/running tasks:", err)
				continue
			}
			for _, t := range ts {
				_, ok := usage[t.RemoteHost]
				if !ok {
					usage[t.RemoteHost] = &node.Usage{}
				}
				usage[t.RemoteHost].Tasks++
				usage[t.RemoteHost].CPUCores += t.Resources.CPUCores
				usage[t.RemoteHost].RAMGb += t.Resources.RAMGb
			}

			nodes, err := m.DB.Active()
			if err != nil {
				log.Println("unable to get active nodes:", err)
				continue
			}

			for _, n := range nodes {
				_, ok := usage[n.Host]
				if !ok {
					usage[n.Host] = &node.Usage{}
				}
				n.Usage = usage[n.Host]
				if err := m.DB.UpdateUsage(n); err != nil {
					log.Printf("unable to update node %s: %v\n", n.Host, err)
				}
			}

			time.Sleep(sleepTime)
		}
	}

	go runQueuedTasks()
	go checkRunningTasks()
	go updateUsage()
}

func (m *Master) initializeTasks() error {
	ts, err := m.DB.FindByState(task.Queued)
	if err != nil {
		return fmt.Errorf("could not get queued tasks: %w", err)
	}
	for _, t := range ts {
		n, err := m.Request(t.Resources)
		switch err.(type) {
		case nil:
			t.RemoteHost = n.Host
			t.State = task.Initializing
			if err := m.DB.Update(t); err != nil {
				log.Println("unable to update task:", err)
			}
			log.Println(t)
			go m.run(t)
		case NoActiveNodes:
			return nil
		case NoEnoughResources:
			continue
		default:
			return err
		}
	}
	return nil
}

func (m *Master) run(t *task.Task) {
	n, err := m.DB.GetByHost(t.RemoteHost)
	if err != nil {
		log.Println("unable to get node by host:", err)
		return
	}
	if !n.Active {
		t.State = task.Queued
	} else {
		if err := m.Runner.Run(t); err != nil {
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
	log.Println(t)
}

func (m *Master) check(t *task.Task) {
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

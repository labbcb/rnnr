package master

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/node"
	"github.com/labbcb/rnnr/task"
)

// NoActiveNodes error is returned when there is no active node for processing tasks remotely.
type NoActiveNodes struct {
	error
}

// NoEnoughResources error is returned when none os active node have enough computing resources to process task.
type NoEnoughResources struct {
	error
}

// Activate adds computing note after requesting its information with available resources.
// If the node is already registered it is activated keeping the previous ID but updating information.
// Usage information is reset.
func (m *Master) Activate(n *node.Node) error {
	info, err := client.GetInfo(n.Host)
	if err != nil {
		return fmt.Errorf("unable to get info of node %s: %w", n.Host, err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("unable to generate node id: %w", err)
	}
	n.ID = id.String()
	n.Info = info
	n.Active = true
	n.Usage = &node.Usage{}
	if err := m.DB.Add(n); err != nil {
		return fmt.Errorf("unable to add node %s, %w", n.Host, err)
	}

	log.Println(n)

	return nil
}

// Deactivate updates node availability removing usage information.
// Also cancels remote tasks and put them to queue.
func (m *Master) Deactivate(id string) error {
	n, err := m.DB.GetByID(id)
	if err != nil {
		return fmt.Errorf("unable to get node %s: %w", id, err)
	}

	n.Active = false
	n.Usage = &node.Usage{}
	if err := m.DB.UpdateNode(n); err != nil {
		return fmt.Errorf("unable to update disabled node %s: %w", id, err)
	}

	ts, err := m.DB.FindByState(task.Initializing, task.Running, task.Paused)
	if err != nil {
		return fmt.Errorf("unable to get tasks from node %s: %w", n.Host, err)
	}
	for _, t := range ts {
		if t.RemoteHost != n.Host {
			continue
		}

		go m.Runner.Cancel(t)
	}

	log.Println(n)

	return nil
}

// All returns all computing node (deactivated included).
func (m *Master) All() ([]*node.Node, error) {
	ns, err := m.DB.All()
	if err != nil {
		return nil, fmt.Errorf("unable to get all nodes: %w", err)
	}
	return ns, nil
}

// Request selects a node that have enough computing resource to execute task.
// If there is no active node it returns NoActiveNodes error.
// If there is some active none but none of them is able to process then it returns NoEnoughResources error.
// Once found a node it will update in database.
func (m *Master) Request(r *task.Resources) (*node.Node, error) {
	nodes, err := m.DB.Active()
	if err != nil {
		return nil, fmt.Errorf("unable to get active nodes: %w", err)
	}
	if len(nodes) == 0 {
		return nil, NoActiveNodes{errors.New("no active node")}
	}

	usage := make(map[string]*node.Usage)
	ts, err := m.DB.FindByState(task.Initializing, task.Running)
	if err != nil {
		return nil, fmt.Errorf("getting initializing/running tasks: %w", err)
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

	var best *node.Node
	for _, n := range nodes {
		_, ok := usage[n.Host]
		if !ok {
			usage[n.Host] = &node.Usage{}
		}
		n.Usage = usage[n.Host]
		if r.CPUCores <= n.AvailableCPUCores() && r.RAMGb <= n.AvailableRAMGb() {
			if best == nil || n.AvailableCPUCores() > best.AvailableCPUCores() && n.AvailableRAMGb() > best.AvailableRAMGb() {
				best = n
			}
		}
	}

	if best == nil {
		return nil, NoEnoughResources{fmt.Errorf("no active node have enough resources: CPU=%d RAM=%.2fGB", r.CPUCores, r.RAMGb)}
	}

	return best, nil
}

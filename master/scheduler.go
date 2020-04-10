package master

import (
	"errors"
	"fmt"

	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
)

// NoActiveNodes error is returned when there is no active node for processing tasks remotely.
type NoActiveNodes struct {
	error
}

// NoEnoughResources error is returned when none os active node have enough computing resources to process task.
type NoEnoughResources struct {
	error
}

// EnableNode inserts or enables a worker node.
func (m *Master) EnableNode(node *models.Node) error {
	maxCPU, maxRAMGb, err := GetNodeResources(node)
	if err != nil {
		return err
	}

	if node.CPUCores == 0 || node.CPUCores > maxCPU {
		node.CPUCores = maxCPU
	}

	if node.RAMGb == 0 || node.RAMGb > maxRAMGb {
		node.RAMGb = maxRAMGb
	}

	node.Active = true
	node.Usage = &models.Usage{}
	if err := m.DB.AddNodes(node); err != nil {
		return err
	}

	return nil
}

// DisableNode updates node availability removing usage information.
// Also cancels remote tasks and put them to queue.
func (m *Master) DisableNode(host string) error {
	node, err := m.DB.GetNode(host)
	if err != nil {
		return err
	}

	node.Active = false
	node.Usage = &models.Usage{}
	if err := m.DB.UpdateNode(node); err != nil {
		return err
	}

	tasks, err := m.DB.FindByState(models.Initializing, models.Running, models.Paused)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.RemoteHost != node.Host {
			continue
		}

		if err := RemoteCancel(task, node); err != nil {
			log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "host": task.RemoteHost, "error": err}).Error("Unable to remotely cancel task.")
		}
		log.WithField("id", task.ID).Info("Task canceled.")
	}
	return nil
}

// GetAllNodes returns all computing node (deactivated included).
func (m *Master) GetAllNodes() ([]*models.Node, error) {
	ns, err := m.DB.AllNodes()
	if err != nil {
		return nil, fmt.Errorf("unable to get all nodes: %w", err)
	}
	return ns, nil
}

// RequestNode selects a node that have enough computing resource to execute task.
// If there is no active node it returns NoActiveNodes error.
// If there is some active node but none of them is able to process then it returns NoEnoughResources error.
// Once found a node it will update in database.
func (m *Master) RequestNode(resources *models.Resources) (*models.Node, error) {
	// GetTask active computing nodes.
	nodes, err := m.DB.GetActiveNodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, NoActiveNodes{errors.New("no active node")}
	}

	// Update workload of active nodes.
	if err := m.UpdateNodesWorkload(nodes); err != nil {
		return nil, fmt.Errorf("unable to update node workload: %w", err)
	}

	// Get the best node for the requested computing resources.
	// The selected node should have enough CPU and Memory available.
	// Calculate how much free resources nodes would have if task is assigned.
	// The node with least free resources available is selected.
	var bestNode *models.Node
	for _, node := range nodes {
		freeCPU := node.CPUCores - node.Usage.CPUCores - resources.CPUCores
		freeRAMGb := node.RAMGb - node.Usage.RAMGb - resources.RAMGb

		if freeCPU < 0 || freeRAMGb < 0 {
			continue
		}

		if bestNode == nil || freeCPU <= bestNode.CPUCores && freeRAMGb <= bestNode.RAMGb {
			bestNode = node
		}
	}

	if bestNode == nil {
		return nil, NoEnoughResources{fmt.Errorf("no active node have enough resources: CPU=%d RAM=%.2fGB", resources.CPUCores, resources.RAMGb)}
	}

	return bestNode, nil
}

// UpdateNodesWorkload gets active tasks (Initializing or Running) and update node usage.
func (m *Master) UpdateNodesWorkload(nodes []*models.Node) error {
	usage := make(map[string]*models.Usage)
	tasks, err := m.DB.FindByState(models.Initializing, models.Running)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		_, ok := usage[task.RemoteHost]
		if !ok {
			usage[task.RemoteHost] = &models.Usage{}
		}
		usage[task.RemoteHost].Tasks++
		usage[task.RemoteHost].CPUCores += task.Resources.CPUCores
		usage[task.RemoteHost].RAMGb += task.Resources.RAMGb
	}

	for _, n := range nodes {
		_, ok := usage[n.Host]
		if !ok {
			usage[n.Host] = &models.Usage{}
		}
		n.Usage = usage[n.Host]
	}

	return nil
}

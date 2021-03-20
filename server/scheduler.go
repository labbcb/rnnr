package server

import (
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
// Use the current computational resources of the node if it had not been defined.
func (m *Main) EnableNode(node *models.Node) error {
	info, err := GetNodeResources(node)
	if err != nil {
		return err
	}

	if node.CPUCores == 0 {
		node.CPUCores = info.CpuCores
	}

	if node.RAMGb == 0 {
		node.RAMGb = info.RamGb
	}

	if node.CPUCores > info.IdentifiedCpuCores {
		log.Warnf("Defined number of CPU cores (%d) is greater than identified (%d).", node.CPUCores, info.IdentifiedCpuCores)
	}

	if node.RAMGb > info.IdentifiedRamGb {
		log.Warnf("Defined number of RAM (%.2f GB) is greater than identified (%.2f GB).", node.RAMGb, info.IdentifiedRamGb)
	}

	node.Active = true
	node.Usage = &models.Usage{}
	if err := m.DB.AddNode(node); err != nil {
		return err
	}

	return nil
}

// DisableNode updates node availability removing usage information.
// Cancel argument will cancels remote tasks and puts them back to queue.
func (m *Main) DisableNode(host string, cancel bool) error {
	node, err := m.DB.GetNode(host)
	if err != nil {
		return err
	}

	node.Active = false
	node.Usage = &models.Usage{}
	if err := m.DB.UpdateNode(node); err != nil {
		return err
	}

	if !cancel {
		return nil
	}

	go func() {
		tasks, err := m.DB.ListTasks(0, 0, models.Full, []string{host}, []models.State{models.Initializing, models.Running, models.Paused})
		if err != nil {
			log.WithError(err).Warn("Unable to get tasks to cancel.")
			return
		}

		for _, task := range tasks {
			if err := RemoteCancel(task, node); err != nil {
				log.WithError(err).WithFields(log.Fields{"id": task.ID, "host": task.Host}).Warn("Unable to remotely cancel task.")
			}

			go m.enqueueTask(task)
		}
	}()

	return nil
}

// ListNodes returns worker nodes (disabled included).
// Set active to return active (enabled) or disable nodes.
func (m *Main) ListNodes(active *bool) ([]*models.Node, error) {
	return m.DB.ListNodes(active)
}

// RequestNode selects a node that have enough computing resource to execute task.
// If there is no active node it returns NoActiveNodes error.
// If there is some active node but none of them is able to process then it returns NoEnoughResources error.
// Once found a node it will update in database.
func (m *Main) RequestNode(resources *models.Resources) (*models.Node, error) {
	// GetTask active computing nodes.
	active := true
	nodes, err := m.DB.ListNodes(&active)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NoActiveNodes{}
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
		return nil, &NoEnoughResources{}
	}

	return bestNode, nil
}

// UpdateNodesWorkload gets active tasks (Initializing or Running) and update node usage.
func (m *Main) UpdateNodesWorkload(nodes []*models.Node) error {
	usage := make(map[string]*models.Usage)
	tasks, err := m.DB.ListTasks(0, 0, models.Full, nil, []models.State{models.Initializing, models.Running})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		_, ok := usage[task.Host]
		if !ok {
			usage[task.Host] = &models.Usage{}
		}
		usage[task.Host].Tasks++
		usage[task.Host].CPUCores += task.Resources.CPUCores
		usage[task.Host].RAMGb += task.Resources.RAMGb
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

func (m *Main) enqueueTask(task *models.Task) {
	task.State = models.Queued
	task.Logs = []*models.Log{{}}
	task.Host = ""
	task.Metrics = nil
	if err := m.DB.UpdateTask(task); err != nil {
		log.WithFields(log.Fields{"id": task.ID, "name": task.Name, "error": err}).Error("Unable to update task.")
		return
	}
	log.WithField("id", task.ID).Info("Task enqueued.")
}

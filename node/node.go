package node

import (
	"fmt"
)

// Node is a computing node that accepts and executes tasks.
// It has maximum allowed (Info) and real-time allocated computing resources (Usage).
// Host is used as its unique identifier.
type Node struct {
	ID     string `json:"id" bson:"_id"`
	Host   string `json:"host"`
	Active bool   `json:"active"`
	Info   *Info  `json:"info"`

	// Usage keeps real-time allocated resources in memory. It is not stored in database.
	Usage *Usage `json:"usage" bson:"-"`
}

// Info has the maximum computing resources.
type Info struct {
	CPUCores int     `json:"cpuCores"`
	RAMGb    float64 `json:"ramGb"`
}

// Usage has the amount of computings resources already in use.
type Usage struct {
	Tasks    int     `json:"tasks"`
	CPUCores int     `json:"cpuCores"`
	RAMGb    float64 `json:"ramGb"`
}

func (n *Node) String() string {
	var msg string
	if n.Active {
		msg = "ACTIVE"
	} else {
		msg = "INACTIVE"
	}
	return fmt.Sprintf("node %s %s CPU=%02d/%02d RAM=%.2f/%.2fGB %d tasks %s",
		n.ID, n.Host, n.Usage.CPUCores, n.Info.CPUCores, n.Usage.RAMGb, n.Info.RAMGb, n.Usage.Tasks, msg)
}

// AvailableCPUCores returns amount of free CPU cores.
func (n *Node) AvailableCPUCores() int {
	return n.Info.CPUCores - n.Usage.CPUCores
}

// AvailableRAMGb returns amount of free RAM memory in GB.
func (n *Node) AvailableRAMGb() float64 {
	return n.Info.RAMGb - n.Usage.RAMGb
}

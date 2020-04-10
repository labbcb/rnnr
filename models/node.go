package models

// Node is a computing node that accepts and executes tasks.
// It has maximum allowed (Info) and real-time allocated computing resources (Usage).
// Host is used as its unique identifier.
type Node struct {
	Host   string `json:"host" bson:"_id"`
	Port   string `json:"port"`
	Active bool   `json:"active"`

	CPUCores int32   `json:"cpuCores"`
	RAMGb    float64 `json:"ramGb"`

	// Usage keeps real-time allocated resources in memory. It is not stored in database.
	Usage *Usage `json:"usage" bson:"-"`
}

// Address returns full node address with port.
func (n *Node) Address() string {
	return n.Host + ":" + n.Port
}

// Usage has the amount of computings resources already in use.
type Usage struct {
	Tasks    int     `json:"tasks"`
	CPUCores int32   `json:"cpuCores"`
	RAMGb    float64 `json:"ramGb"`
}

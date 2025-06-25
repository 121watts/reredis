package cluster

// Slot represents a range of hash slots assigned to a cluster node.
// This enables horizontal data partitioning by dividing the key space
// into manageable chunks that can be redistributed during cluster changes.
type Slot struct {
	Start int32
	End   int32
}

// Node represents a single server instance in the cluster.
// Each node maintains its identity, network location, and assigned slot range
// to enable distributed operations and cluster management.
type Node struct {
	ID   string // Unique identifier for cluster communication and consistency
	Slot Slot   // Hash slot range this node is responsible for
	Host string // Network address for client and inter-node communication  
	Port string // Network port for establishing connections
}

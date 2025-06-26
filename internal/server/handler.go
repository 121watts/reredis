package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/121watts/reredis/internal/cluster"
	"github.com/121watts/reredis/internal/observer"
	"github.com/121watts/reredis/internal/store"
	"github.com/121watts/reredis/internal/wal"
)

type CommandHandler struct {
	store          *store.Store
	hub            *observer.Hub
	walWriter      *wal.Writer
	clusterManager *cluster.Manager
	logger         *slog.Logger
}

func NewCommandHandler(store *store.Store, hub *observer.Hub, ww *wal.Writer, cm *cluster.Manager, logger *slog.Logger) *CommandHandler {
	return &CommandHandler{
		store:          store,
		hub:            hub,
		walWriter:      ww,
		clusterManager: cm,
		logger:         logger,
	}
}

type OperationResult struct {
	Key        string
	Value      string
	Action     string
	NeedsStats bool
}

func (c *CommandHandler) HandleSet(parts []string) (*OperationResult, error) {
	const expectedParts = 3
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("wrong number of arguments for 'SET'")
	}

	k, v := parts[1], parts[2]

	// Write to WAL first
	err := c.walWriter.WriteCommand(parts)
	if err != nil {
		c.logger.Error("failed to write to WAL", "error", err)
		return nil, fmt.Errorf("failed to write to WAL: %w", err)
	}

	// Check if this is a new key to update cluster statistics
	oldValue, existsBefore := c.store.Get(k)

	c.store.Set(k, v)

	// Update cluster statistics
	if c.clusterManager != nil {
		if !existsBefore {
			// New key: increment count and add byte size
			c.clusterManager.IncrementKeyCount()
			c.clusterManager.AddByteSize(len(k), len(v))
		} else {
			// Existing key: update byte size (subtract old, add new)
			c.clusterManager.SubtractByteSize(len(k), len(oldValue))
			c.clusterManager.AddByteSize(len(k), len(v))
		}
	}

	return &OperationResult{
		Key:        k,
		Value:      v,
		Action:     "set",
		NeedsStats: c.clusterManager != nil && len(c.clusterManager.Nodes) > 1,
	}, nil
}

func (c *CommandHandler) HandleGet(parts []string) (string, error) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		return "", fmt.Errorf("wrong number of arguments for 'GET'")
	}

	k := parts[1]
	v, ok := c.store.Get(k)

	if !ok {
		return "", fmt.Errorf("key not found")
	}

	return v, nil
}

func (c *CommandHandler) HandleDelete(parts []string) (bool, *OperationResult, error) {
	const expectedParts = 2

	if len(parts) != expectedParts {
		return false, nil, fmt.Errorf("wrong number of arguments for 'DEL'")
	}

	k := parts[1]

	// Write to WAL first
	err := c.walWriter.WriteCommand(parts)
	if err != nil {
		c.logger.Error("failed to write to WAL", "error", err)
		return false, nil, fmt.Errorf("failed to write to WAL: %w", err)
	}

	// Get the value before deletion to track byte size
	oldValue, existed := c.store.Get(k)
	ok := c.store.Delete(k)

	if ok {
		// Update cluster statistics when key is deleted
		if c.clusterManager != nil {
			c.clusterManager.DecrementKeyCount()
			if existed {
				c.clusterManager.SubtractByteSize(len(k), len(oldValue))
			}
		}

		return true, &OperationResult{
			Key:        k,
			Value:      "",
			Action:     "del",
			NeedsStats: c.clusterManager != nil && len(c.clusterManager.Nodes) > 1,
		}, nil
	}

	return false, nil, nil
}

// checkSlotOwnership returns the key for the given key, or nil if current node owns it
func (c *CommandHandler) checkSlotOwnership(key string) string {
	// Defensive check: if cluster manager is nil, allow all operations
	if c.clusterManager == nil {
		return ""
	}

	// If cluster is not initialized (< 3 nodes), current node handles all slots
	if len(c.clusterManager.Nodes) < 3 {
		return ""
	}

	// Defensive check: if current node is nil, allow all operations
	if c.clusterManager.Node == nil {
		return ""
	}

	slot := cluster.CalculateSlot(key)
	node := c.clusterManager.Node

	if slot < node.Slot.Start || slot > node.Slot.End {
		ownerNode := c.clusterManager.GetNodeForSlots(slot)
		if ownerNode != nil {
			// Return the key for MOVED response
			return key
		}
	}
	return ""
}

func (c *CommandHandler) HandleCluster(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("wrong number of arguments for 'CLUSTER'")
	}

	subcommand := strings.ToUpper(parts[1])

	switch subcommand {
	case "MEET":
		return c.handleClusterMeet(parts)
	case "NODES":
		return c.handleClusterNodes(parts)
	case "INFO":
		return c.handleClusterInfo(parts)
	default:
		return fmt.Errorf("unknown cluster subcommand '%s'", subcommand)
	}
}

func (c *CommandHandler) handleClusterMeet(parts []string) error {
	if len(parts) != 4 {
		return fmt.Errorf("wrong number of arguments for 'CLUSTER MEET'")
	}

	host := parts[2]
	port := parts[3]

	c.logger.Info("cluster meet requested", "ip", host, "port", port)

	// Add a new node to the cluster (simplified - normally you'd connect and exchange info)
	// This enables cluster discovery and growth by building the node topology
	c.clusterManager.AddNode(host, port)

	c.logger.Info("node added to cluster", "total-nodes", len(c.clusterManager.Nodes))

	return nil
}

func (c *CommandHandler) handleClusterNodes(parts []string) error {
	// TODO: Return cluster topology information
	return fmt.Errorf("TODO: cluster nodes not implemented")
}

func (c *CommandHandler) handleClusterInfo(parts []string) error {
	// TODO: Return cluster status information
	return fmt.Errorf("TODO: cluster info not implemented")
}

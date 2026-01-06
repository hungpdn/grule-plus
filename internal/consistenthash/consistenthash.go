package consistenthash

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// HashFunc defines the function to hash keys
type HashFunc func(data []byte) uint32

// ConsistentHash represents a consistent hashing ring
type ConsistentHash struct {
	hashFunc HashFunc          // hash function
	replicas int               // number of virtual nodes per real node
	keys     []uint32          // sorted hash ring
	nodeMap  map[uint32]string // hash -> node
	nodes    map[string]bool   // real nodes
	mu       sync.RWMutex      // mutex for thread safety
}

// New creates a new ConsistentHash instance
func New(replicas int, hashFunc HashFunc) *ConsistentHash {
	if hashFunc == nil {
		hashFunc = md5Hash
	}
	return &ConsistentHash{
		hashFunc: hashFunc,
		replicas: replicas,
		nodeMap:  make(map[uint32]string),
		nodes:    make(map[string]bool),
	}
}

// md5Hash is the default hash function using MD5
func md5Hash(data []byte) uint32 {
	hash := md5.Sum(data)
	return uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
}

// AddNode adds a node to the hash ring
func (c *ConsistentHash) AddNode(node string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.nodes[node] {
		return // node already exists
	}
	c.nodes[node] = true

	// Add virtual nodes
	for i := 0; i < c.replicas; i++ {
		hash := c.hashFunc([]byte(node + strconv.Itoa(i)))
		c.keys = append(c.keys, hash)
		c.nodeMap[hash] = node
	}

	// Sort the keys
	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})
}

// RemoveNode removes a node from the hash ring
func (c *ConsistentHash) RemoveNode(node string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.nodes[node] {
		return // node doesn't exist
	}
	delete(c.nodes, node)

	// Remove virtual nodes
	newKeys := make([]uint32, 0, len(c.keys))
	for _, key := range c.keys {
		if c.nodeMap[key] != node {
			newKeys = append(newKeys, key)
		} else {
			delete(c.nodeMap, key)
		}
	}
	c.keys = newKeys
}

// GetNode returns the node responsible for the given key
func (c *ConsistentHash) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.keys) == 0 {
		return ""
	}

	hash := c.hashFunc([]byte(key))

	// Find the first node with hash >= key hash
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	// If no node found with hash >= key hash, wrap around to first node
	if idx == len(c.keys) {
		idx = 0
	}

	return c.nodeMap[c.keys[idx]]
}

// GetNodes returns all nodes in the ring
func (c *ConsistentHash) GetNodes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make([]string, 0, len(c.nodes))
	for node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodeCount returns the number of real nodes
func (c *ConsistentHash) GetNodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes)
}

// GetVirtualNodeCount returns the total number of virtual nodes
func (c *ConsistentHash) GetVirtualNodeCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.keys)
}

// IsEmpty returns true if the hash ring has no nodes
func (c *ConsistentHash) IsEmpty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.nodes) == 0
}

// GetNodeStats returns statistics about node distribution
func (c *ConsistentHash) GetNodeStats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]int)
	for _, node := range c.nodeMap {
		stats[node]++
	}
	return stats
}

// String returns a string representation of the consistent hash
func (c *ConsistentHash) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("ConsistentHash{nodes=%d, virtual_nodes=%d, replicas=%d}",
		len(c.nodes), len(c.keys), c.replicas)
}

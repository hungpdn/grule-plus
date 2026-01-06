package consistenthash

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ch := New(3, nil)
	assert.NotNil(t, ch)
	assert.Equal(t, 3, ch.replicas)
	assert.NotNil(t, ch.hashFunc)
	assert.NotNil(t, ch.nodeMap)
	assert.NotNil(t, ch.nodes)
	assert.Empty(t, ch.keys)
}

func TestAddNode(t *testing.T) {
	ch := New(2, nil)

	// Add first node
	ch.AddNode("node1")
	assert.True(t, ch.nodes["node1"])
	assert.Len(t, ch.keys, 2) // 2 replicas
	assert.Len(t, ch.nodeMap, 2)

	// Add second node
	ch.AddNode("node2")
	assert.True(t, ch.nodes["node2"])
	assert.Len(t, ch.keys, 4) // 4 total virtual nodes
	assert.Len(t, ch.nodeMap, 4)

	// Add duplicate node (should not change anything)
	ch.AddNode("node1")
	assert.Len(t, ch.keys, 4)
	assert.Len(t, ch.nodeMap, 4)
}

func TestRemoveNode(t *testing.T) {
	ch := New(2, nil)

	// Add nodes
	ch.AddNode("node1")
	ch.AddNode("node2")
	assert.Len(t, ch.keys, 4)

	// Remove existing node
	ch.RemoveNode("node1")
	assert.False(t, ch.nodes["node1"])
	assert.True(t, ch.nodes["node2"])
	assert.Len(t, ch.keys, 2) // only node2's virtual nodes remain

	// Remove non-existing node (should not change anything)
	ch.RemoveNode("node3")
	assert.Len(t, ch.keys, 2)
}

func TestGetNode(t *testing.T) {
	ch := New(3, nil)

	// Empty ring
	assert.Empty(t, ch.GetNode("key1"))

	// Add nodes
	ch.AddNode("node1")
	ch.AddNode("node2")
	ch.AddNode("node3")

	// Test key distribution
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	nodeCounts := make(map[string]int)

	for _, key := range keys {
		node := ch.GetNode(key)
		assert.NotEmpty(t, node)
		assert.Contains(t, []string{"node1", "node2", "node3"}, node)
		nodeCounts[node]++
	}

	// Should have some distribution (not all keys on one node)
	assert.True(t, len(nodeCounts) > 1, "Keys should be distributed across multiple nodes")
}

func TestGetNodeDeterministic(t *testing.T) {
	ch := New(3, nil)
	ch.AddNode("node1")
	ch.AddNode("node2")

	// Same key should always return same node
	key := "test-key"
	node1 := ch.GetNode(key)
	node2 := ch.GetNode(key)
	assert.Equal(t, node1, node2)
}

func TestGetNodes(t *testing.T) {
	ch := New(2, nil)

	// Empty
	nodes := ch.GetNodes()
	assert.Empty(t, nodes)

	// Add nodes
	ch.AddNode("node1")
	ch.AddNode("node2")
	nodes = ch.GetNodes()
	assert.Len(t, nodes, 2)
	assert.Contains(t, nodes, "node1")
	assert.Contains(t, nodes, "node2")
}

func TestGetNodeCount(t *testing.T) {
	ch := New(2, nil)
	assert.Equal(t, 0, ch.GetNodeCount())

	ch.AddNode("node1")
	assert.Equal(t, 1, ch.GetNodeCount())

	ch.AddNode("node2")
	assert.Equal(t, 2, ch.GetNodeCount())

	ch.RemoveNode("node1")
	assert.Equal(t, 1, ch.GetNodeCount())
}

func TestGetVirtualNodeCount(t *testing.T) {
	ch := New(3, nil)
	assert.Equal(t, 0, ch.GetVirtualNodeCount())

	ch.AddNode("node1") // 3 virtual nodes
	assert.Equal(t, 3, ch.GetVirtualNodeCount())

	ch.AddNode("node2") // 6 total virtual nodes
	assert.Equal(t, 6, ch.GetVirtualNodeCount())
}

func TestIsEmpty(t *testing.T) {
	ch := New(2, nil)
	assert.True(t, ch.IsEmpty())

	ch.AddNode("node1")
	assert.False(t, ch.IsEmpty())

	ch.RemoveNode("node1")
	assert.True(t, ch.IsEmpty())
}

func TestGetNodeStats(t *testing.T) {
	ch := New(2, nil)
	ch.AddNode("node1")
	ch.AddNode("node2")

	stats := ch.GetNodeStats()
	assert.Len(t, stats, 2)
	assert.Equal(t, 2, stats["node1"]) // 2 virtual nodes
	assert.Equal(t, 2, stats["node2"]) // 2 virtual nodes
}

func TestString(t *testing.T) {
	ch := New(3, nil)
	str := ch.String()
	assert.Contains(t, str, "ConsistentHash")
	assert.Contains(t, str, "nodes=0")
	assert.Contains(t, str, "virtual_nodes=0")
	assert.Contains(t, str, "replicas=3")

	ch.AddNode("node1")
	str = ch.String()
	assert.Contains(t, str, "nodes=1")
	assert.Contains(t, str, "virtual_nodes=3")
}

func TestNodeRemovalDistribution(t *testing.T) {
	ch := New(3, nil)

	// Add many nodes
	for i := 1; i <= 10; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	// Generate many keys and record their nodes
	keyNodes := make(map[string]string)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		node := ch.GetNode(key)
		keyNodes[key] = node
	}

	// Remove one node
	ch.RemoveNode("node5")

	// Check how many keys changed nodes (should be minimal)
	changed := 0
	total := 0
	for key, oldNode := range keyNodes {
		newNode := ch.GetNode(key)
		if newNode != oldNode {
			changed++
		}
		total++
	}

	// Less than 15% of keys should change (rough estimate for good distribution)
	changeRatio := float64(changed) / float64(total)
	assert.True(t, changeRatio < 0.15, "Too many keys changed nodes: %d/%d (%.2f%%)", changed, total, changeRatio*100)
}

func TestConcurrentAccess(t *testing.T) {
	ch := New(2, nil)

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			ch.AddNode(fmt.Sprintf("node%d", i))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			ch.GetNode(fmt.Sprintf("key%d", i))
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic and should have some nodes
	assert.True(t, ch.GetNodeCount() > 0)
}

func TestCustomHashFunc(t *testing.T) {
	// Simple hash function for testing
	customHash := func(data []byte) uint32 {
		sum := uint32(0)
		for _, b := range data {
			sum += uint32(b)
		}
		return sum
	}

	ch := New(2, customHash)
	ch.AddNode("node1")
	ch.AddNode("node2")

	// Test that custom hash function is used
	node := ch.GetNode("test")
	assert.NotEmpty(t, node)
}

func BenchmarkAddNode(b *testing.B) {
	ch := New(10, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}
}

func BenchmarkGetNode(b *testing.B) {
	ch := New(3, nil)
	for i := 0; i < 10; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.GetNode(fmt.Sprintf("key%d", i))
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	ch := New(3, nil)
	for i := 0; i < 100; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.RemoveNode(fmt.Sprintf("node%d", i%100))
	}
}

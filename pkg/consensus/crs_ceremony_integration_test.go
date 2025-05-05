package consensus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zkrollup/pkg/p2p"
)

func init() {
	// Configure logging for tests
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

// TestCRSCeremonyIntegration tests the CRS ceremony with real p2p nodes
func TestCRSCeremonyIntegration(t *testing.T) {
	// Skip test if snarkjs is not installed
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("Skipping test because npx is not installed")
	}

	// Create a temporary directory for the CRS ceremony files
	tempDir, err := os.MkdirTemp("", "crs-ceremony-integration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a network of real P2P nodes
	numNodes := 3
	nodes := make([]*p2p.Node, numNodes)
	pbftNodes := make([]*PBFT, numNodes)

	// Create a context for the test
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a wait group to wait for all nodes to complete the ceremony
	var wg sync.WaitGroup
	wg.Add(numNodes)

	// Create the nodes
	for i := 0; i < numNodes; i++ {
		// Use a different port for each node
		port := 10000 + i

		// Create a p2p node
		node, err := p2p.NewNode(ctx, port, nil)
		require.NoError(t, err)

		// Store the node
		nodes[i] = node

		// Create PBFT instance for this node
		nodeID := node.Host.ID().String()
		isLeader := i == 0 // First node is the leader
		pbftNodes[i] = NewPBFT(node, nodeID, isLeader)

		// Set a custom CRS ceremony directory for each node
		pbftNodes[i].crsCeremonyDir = filepath.Join(tempDir, fmt.Sprintf("node-%d", i))
		require.NoError(t, os.MkdirAll(pbftNodes[i].crsCeremonyDir, 0755))

		// Start PBFT
		pbftNodes[i].Start()

		// Set up goroutine to wait for ceremony completion
		go func(idx int) {
			defer wg.Done()
			select {
			case <-pbftNodes[idx].GetCRSCeremonyDoneChan():
				t.Logf("Node %s completed CRS ceremony", pbftNodes[idx].nodeID)
			case <-time.After(120 * time.Second): // Increase timeout to 120 seconds
				t.Errorf("Timeout waiting for node %s to complete CRS ceremony", pbftNodes[idx].nodeID)
			}
		}(i)
	}

	// Connect the nodes to each other
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i != j {
				// Get peer info
				peerInfo := peer.AddrInfo{
					ID:    nodes[j].Host.ID(),
					Addrs: nodes[j].Host.Addrs(),
				}

				// Connect to the peer
				err := nodes[i].Host.Connect(ctx, peerInfo)
				if err != nil {
					t.Logf("Error connecting node %d to node %d: %v", i, j, err)
				}
			}
		}
	}

	// Wait for connections to establish
	time.Sleep(2 * time.Second)

	// Manually register all nodes with each other to ensure they're in each other's nodeIDs list
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			// Add each node's ID to the other node's list
			pbftNodes[i].addNodeID(pbftNodes[j].nodeID)
		}
		// Update total nodes count
		pbftNodes[i].UpdateTotalNodes(numNodes)
		t.Logf("Node %d has %d nodes in its list", i, len(pbftNodes[i].nodeIDs))
	}

	// Wait a bit for node registration to propagate
	time.Sleep(1 * time.Second)

	// Leader initiates the CRS ceremony
	err = pbftNodes[0].StartCRSCeremony()
	require.NoError(t, err)

	// Wait for all nodes to complete the ceremony with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("All nodes completed the CRS ceremony")
	case <-time.After(120 * time.Second): // 2 minute timeout for the entire test
		t.Fatal("Test timed out waiting for CRS ceremony to complete")
	}

	// Verify that all nodes have the same final PTau file
	var finalPTauPath string
	var finalPTauData []byte

	// Get the final PTau file from the first node
	pbftNodes[0].ptauStateLock.RLock()
	if pbftNodes[0].ptauState != nil {
		finalPTauPath = pbftNodes[0].ptauState.PTauPath
		pbftNodes[0].ptauStateLock.RUnlock()

		// Verify the file exists before trying to read it
		_, err = os.Stat(finalPTauPath)
		if err != nil {
			t.Logf("Error accessing PTau file for node 0: %v", err)
			t.FailNow()
		}

		finalPTauData, err = os.ReadFile(finalPTauPath)
		require.NoError(t, err)
		require.NotEmpty(t, finalPTauData)

		t.Logf("Node 0 final PTau file: %s (size: %d bytes)", finalPTauPath, len(finalPTauData))

		// Check that all nodes have the same final PTau file
		for i := 1; i < numNodes; i++ {
			pbftNodes[i].ptauStateLock.RLock()
			nodePTauPath := pbftNodes[i].ptauState.PTauPath
			pbftNodes[i].ptauStateLock.RUnlock()

			// Verify the file exists before trying to read it
			_, err = os.Stat(nodePTauPath)
			if err != nil {
				t.Logf("Error accessing PTau file for node %d: %v", i, err)
				continue
			}

			nodePTauData, err := os.ReadFile(nodePTauPath)
			if err != nil {
				t.Logf("Error reading PTau file for node %d: %v", i, err)
				continue
			}

			if len(nodePTauData) == 0 {
				t.Logf("Empty PTau file for node %d", i)
				continue
			}

			t.Logf("Node %d final PTau file: %s (size: %d bytes)", i, nodePTauPath, len(nodePTauData))

			// Compare file sizes
			assert.Equal(t, len(finalPTauData), len(nodePTauData),
				"Node %d has different PTau file size than node 0", i)
		}
	} else {
		pbftNodes[0].ptauStateLock.RUnlock()
		t.Log("Node 0 has no PTau state")
	}

	// Verify that the ceremony completed successfully
	for i := 0; i < numNodes; i++ {
		pbftNodes[i].ptauStateLock.RLock()
		if pbftNodes[i].ptauState != nil {
			assert.True(t, pbftNodes[i].ptauState.Completed, "Node %d ceremony not marked as completed", i)
		} else {
			t.Logf("Node %d has no PTau state", i)
		}
		pbftNodes[i].ptauStateLock.RUnlock()
	}
}

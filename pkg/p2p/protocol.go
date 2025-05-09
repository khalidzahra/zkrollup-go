package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"zkrollup/pkg/state"
)

const (
	// Protocol IDs
	TransactionProtocolID = protocol.ID("/zkrollup/tx/1.0.0")
	BatchProtocolID       = protocol.ID("/zkrollup/batch/1.0.0")
	ConsensusProtocolID   = protocol.ID("/zkrollup/consensus/1.0.0")
)

// Message types
type MessageType int

const (
	MessageTransaction MessageType = iota
	MessageBatch
	MessageConsensus
)

// Message represents a P2P network message
type Message struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}

// Protocol handlers
type ProtocolHandlers struct {
	OnTransaction func(tx *state.Transaction) error
	OnBatch       func(batch *state.Batch) error
	OnConsensus   func(msg []byte) error
}

// Protocol handlers are stored in the Node struct

// SetupProtocols sets up protocol handlers for the node
func (n *Node) SetupProtocols(handlers *ProtocolHandlers) {
	// Store the handlers in the node
	if handlers != nil {
		// Debug the handlers being set
		fmt.Printf("Setting up handlers - OnTransaction: %v, OnBatch: %v, OnConsensus: %v\n",
			handlers.OnTransaction != nil, handlers.OnBatch != nil, handlers.OnConsensus != nil)

		n.handlersLock.Lock()
		n.handlers = handlers
		n.handlersLock.Unlock()

		fmt.Printf("Setting up protocol handlers with non-nil handlers for node %s\n", n.Host.ID().String())
	}
	fmt.Printf("Setting up protocol handlers for node %s\n", n.Host.ID().String())

	n.Host.RemoveStreamHandler(TransactionProtocolID)

	n.Host.SetStreamHandler(TransactionProtocolID, func(s network.Stream) {
		fmt.Printf("Received transaction stream from %s\n", s.Conn().RemotePeer().String())
		defer s.Close()

		// Set read deadline
		s.SetReadDeadline(time.Now().Add(time.Second * 30)) // Increased timeout

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding transaction message: %v\n", err)
			return
		}

		if msg.Type != MessageTransaction {
			fmt.Printf("Invalid message type for transaction protocol: %d\n", msg.Type)
			return
		}

		var tx state.Transaction
		if err := json.Unmarshal(msg.Payload, &tx); err != nil {
			fmt.Printf("Error unmarshaling transaction: %v\n", err)
			return
		}

		fmt.Printf("Successfully decoded transaction from %s to %s\n",
			fmt.Sprintf("%x", tx.From), fmt.Sprintf("%x", tx.To))

		nonceStr := fmt.Sprintf("%d", tx.Nonce)
		fmt.Printf("Using nonce string format '%s' for consistent hash computation\n", nonceStr)

		// Check if we have a transaction handler registered
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		// Debug the handlers
		fmt.Printf("Transaction handler check - handlers: %v, OnTransaction: %v\n",
			handlers != nil, handlers != nil && handlers.OnTransaction != nil)

		if handlers == nil || handlers.OnTransaction == nil {
			fmt.Printf("No transaction handler registered in node handlers for node %s\n", n.Host.ID().String())
			return
		}

		// Call the transaction handler
		fmt.Printf("Calling transaction handler\n")
		if err := handlers.OnTransaction(&tx); err != nil {
			fmt.Printf("Error handling transaction: %v\n", err)
			return
		}

		fmt.Printf("Transaction handled successfully\n")
	})

	fmt.Printf("Transaction protocol handler registered for %s\n", TransactionProtocolID)

	// Batch protocol handler
	n.Host.RemoveStreamHandler(BatchProtocolID)
	n.Host.SetStreamHandler(BatchProtocolID, func(s network.Stream) {
		fmt.Printf("Received batch stream from %s\n", s.Conn().RemotePeer().String())
		defer s.Close()

		// Set read deadline
		s.SetReadDeadline(time.Now().Add(time.Second * 10))

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding batch message: %v\n", err)
			return
		}

		if msg.Type != MessageBatch {
			fmt.Printf("Invalid message type for batch protocol: %d\n", msg.Type)
			return
		}

		var batch state.Batch
		if err := json.Unmarshal(msg.Payload, &batch); err != nil {
			fmt.Printf("Error unmarshaling batch: %v\n", err)
			return
		}

		fmt.Printf("Successfully decoded batch with %d transactions\n", len(batch.Transactions))

		// Check if we have a batch handler registered
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		if handlers == nil || handlers.OnBatch == nil {
			fmt.Printf("No batch handler registered in node handlers\n")
			return
		}

		// Call the batch handler
		fmt.Printf("Calling batch handler\n")
		if err := handlers.OnBatch(&batch); err != nil {
			fmt.Printf("Error handling batch: %v\n", err)
			return
		}

		fmt.Printf("Batch handled successfully\n")
	})

	fmt.Printf("Batch protocol handler registered for %s\n", BatchProtocolID)

	n.Host.RemoveStreamHandler(ConsensusProtocolID)
	n.Host.SetStreamHandler(ConsensusProtocolID, func(s network.Stream) {
		fmt.Printf("Received consensus stream from %s\n", s.Conn().RemotePeer().String())
		defer s.Close()

		// Set read deadline
		s.SetReadDeadline(time.Now().Add(time.Second * 10))

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding consensus message: %v\n", err)
			return
		}

		if msg.Type != MessageConsensus {
			fmt.Printf("Invalid message type for consensus protocol: %d\n", msg.Type)
			return
		}

		fmt.Printf("Successfully decoded consensus message of size %d bytes\n", len(msg.Payload))

		// Check if we have a consensus handler registered
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		if handlers == nil || handlers.OnConsensus == nil {
			fmt.Printf("No consensus handler registered in node handlers\n")
			return
		}

		// Call the consensus handler
		fmt.Printf("Calling consensus handler\n")
		if err := handlers.OnConsensus(msg.Payload); err != nil {
			fmt.Printf("Error handling consensus message: %v\n", err)
			return
		}

		fmt.Printf("Consensus message handled successfully\n")
	})

	fmt.Printf("Consensus protocol handler registered for %s\n", ConsensusProtocolID)
}

// BroadcastTransaction broadcasts a transaction to all connected peers
func (n *Node) BroadcastTransaction(ctx context.Context, tx *state.Transaction) error {
	payload, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	msg := Message{
		Type:    MessageTransaction,
		Payload: payload,
	}

	return n.broadcast(ctx, TransactionProtocolID, msg)
}

// BroadcastBatch broadcasts a batch to all connected peers
func (n *Node) BroadcastBatch(ctx context.Context, batch *state.Batch) error {
	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %v", err)
	}

	msg := Message{
		Type:    MessageBatch,
		Payload: payload,
	}

	return n.broadcast(ctx, BatchProtocolID, msg)
}

// BroadcastConsensus broadcasts a consensus message to all connected peers
func (n *Node) BroadcastConsensus(ctx context.Context, payload []byte) error {
	msg := Message{
		Type:    MessageConsensus,
		Payload: payload,
	}

	// Log the broadcast for debugging
	fmt.Printf("Broadcasting consensus message of size %d bytes\n", len(payload))

	// Ensure we have peers to broadcast to
	peers := n.GetPeers()
	if len(peers) == 0 {
		fmt.Printf("No peers available for consensus broadcast\n")
	}

	return n.broadcast(ctx, ConsensusProtocolID, msg)
}

// broadcast sends a message to all connected peers
func (n *Node) broadcast(ctx context.Context, protocolID protocol.ID, msg Message) error {
	peers := n.GetPeers()
	if len(peers) == 0 {
		fmt.Printf("No peers available for broadcasting protocol %s, continuing in standalone mode\n", protocolID)
		return nil
	}

	fmt.Printf("Broadcasting to %d peers using protocol %s\n", len(peers), protocolID)

	successCount := 0
	for _, peer := range peers {
		// Try up to 5 times with a short delay between attempts
		var stream network.Stream
		var err error
		var succeeded bool

		for attempts := 0; attempts < 5; attempts++ {
			// Create a stream with a timeout to avoid hanging
			streamCtx, cancel := context.WithTimeout(ctx, time.Second*30) // Increased timeout for protocol negotiation
			fmt.Printf("Attempt %d: Creating stream to peer %s for protocol %s\n", attempts+1, peer.String(), protocolID)
			stream, err = n.Host.NewStream(streamCtx, peer, protocolID)
			cancel()

			if err == nil {
				succeeded = true
				break
			}

			fmt.Printf("Attempt %d: Failed to create stream to peer %s for protocol %s: %v\n",
				attempts+1, peer.String(), protocolID, err)

			// Only retry if this looks like a protocol negotiation issue
			if attempts < 2 && (err.Error() == "failed to negotiate protocol: protocols not supported" ||
				err.Error() == "protocol not supported") {
				fmt.Printf("Protocol negotiation issue detected, retrying after delay...\n")
				time.Sleep(time.Millisecond * 500)
			} else {
				break
			}
		}

		if !succeeded {
			fmt.Printf("Failed to create stream to peer %s after retries\n", peer.String())
			continue
		}

		// Set deadline for writing to stream
		stream.SetWriteDeadline(time.Now().Add(time.Second * 10)) // Increased timeout

		if err := json.NewEncoder(stream).Encode(msg); err != nil {
			fmt.Printf("Failed to send message to peer %s: %v\n", peer.String(), err)
			stream.Close()
			continue
		}

		stream.Close()
		successCount++
		fmt.Printf("Successfully sent message to peer %s\n", peer.String())
	}

	fmt.Printf("Broadcast summary: %d/%d successful\n", successCount, len(peers))

	if successCount == 0 && len(peers) > 0 {
		return fmt.Errorf("failed to broadcast message to any of the %d peers", len(peers))
	}

	return nil
}

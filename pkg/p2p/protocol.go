package p2p

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"zkrollup/pkg/state"
)

const (
	// Protocol IDs
	TransactionProtocolID = protocol.ID("/zkrollup/tx/1.0.0")
	BatchProtocolID      = protocol.ID("/zkrollup/batch/1.0.0")
	ConsensusProtocolID  = protocol.ID("/zkrollup/consensus/1.0.0")
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
	Type    MessageType     `json:"type"`
	Payload []byte         `json:"payload"`
}

// Protocol handlers
type ProtocolHandlers struct {
	OnTransaction func(tx *state.Transaction) error
	OnBatch      func(batch *state.Batch) error
	OnConsensus  func(msg []byte) error
}

// SetupProtocols sets up protocol handlers for the node
func (n *Node) SetupProtocols(handlers *ProtocolHandlers) {
	// Transaction protocol handler
	n.Host.SetStreamHandler(TransactionProtocolID, func(s network.Stream) {
		defer s.Close()

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding transaction message: %v\n", err)
			return
		}

		if msg.Type != MessageTransaction {
			fmt.Printf("Invalid message type for transaction protocol\n")
			return
		}

		var tx state.Transaction
		if err := json.Unmarshal(msg.Payload, &tx); err != nil {
			fmt.Printf("Error unmarshaling transaction: %v\n", err)
			return
		}

		if err := handlers.OnTransaction(&tx); err != nil {
			fmt.Printf("Error handling transaction: %v\n", err)
			return
		}
	})

	// Batch protocol handler
	n.Host.SetStreamHandler(BatchProtocolID, func(s network.Stream) {
		defer s.Close()

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding batch message: %v\n", err)
			return
		}

		if msg.Type != MessageBatch {
			fmt.Printf("Invalid message type for batch protocol\n")
			return
		}

		var batch state.Batch
		if err := json.Unmarshal(msg.Payload, &batch); err != nil {
			fmt.Printf("Error unmarshaling batch: %v\n", err)
			return
		}

		if err := handlers.OnBatch(&batch); err != nil {
			fmt.Printf("Error handling batch: %v\n", err)
			return
		}
	})

	// Consensus protocol handler
	n.Host.SetStreamHandler(ConsensusProtocolID, func(s network.Stream) {
		defer s.Close()

		var msg Message
		if err := json.NewDecoder(s).Decode(&msg); err != nil {
			fmt.Printf("Error decoding consensus message: %v\n", err)
			return
		}

		if msg.Type != MessageConsensus {
			fmt.Printf("Invalid message type for consensus protocol\n")
			return
		}

		if err := handlers.OnConsensus(msg.Payload); err != nil {
			fmt.Printf("Error handling consensus message: %v\n", err)
			return
		}
	})
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

	return n.broadcast(ctx, ConsensusProtocolID, msg)
}

// broadcast sends a message to all connected peers
func (n *Node) broadcast(ctx context.Context, protocolID protocol.ID, msg Message) error {
	peers := n.GetPeers()
	for _, peer := range peers {
		stream, err := n.Host.NewStream(ctx, peer, protocolID)
		if err != nil {
			fmt.Printf("Failed to create stream to peer %s: %v\n", peer.String(), err)
			continue
		}
		defer stream.Close()

		if err := json.NewEncoder(stream).Encode(msg); err != nil {
			fmt.Printf("Failed to send message to peer %s: %v\n", peer.String(), err)
			continue
		}
	}

	return nil
}

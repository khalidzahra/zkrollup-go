package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a P2P node in the network
type Node struct {
	Host        host.Host
	PingService *ping.PingService
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, port int) (*Node, error) {
	// Create multiaddr for listening
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to create multiaddr: %v", err)
	}

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	// Create ping service for node discovery and health checks
	ps := ping.NewPingService(h)

	// Print node info
	fmt.Printf("Node started with ID: %s\n", h.ID().String())
	fmt.Printf("Node addresses:\n")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID().String())
	}

	return &Node{
		Host:        h,
		PingService: ps,
	}, nil
}

// Connect connects to a peer using their multiaddr
func (n *Node) Connect(ctx context.Context, peerAddr string) error {
	// Parse the peer multiaddr
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %v", err)
	}

	// Extract the peer ID from the multiaddr
	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to get peer info: %v", err)
	}

	// Connect to the peer
	if err := n.Host.Connect(ctx, *info); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	fmt.Printf("Connected to peer: %s\n", info.ID.String())
	return nil
}

// Disconnect from a peer
func (n *Node) Disconnect(ctx context.Context, peerID peer.ID) error {
	if err := n.Host.Network().ClosePeer(peerID); err != nil {
		return fmt.Errorf("failed to disconnect from peer: %v", err)
	}
	return nil
}

// GetPeers returns a list of connected peers
func (n *Node) GetPeers() []peer.ID {
	return n.Host.Network().Peers()
}

// Close shuts down the node
func (n *Node) Close() error {
	return n.Host.Close()
}

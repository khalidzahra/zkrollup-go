package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

const (
	DiscoveryNamespace = "zkrollup"
	DiscoveryInterval  = time.Second * 1
)

// Node represents a P2P node in the network
type Node struct {
	Host            host.Host
	PingService     *ping.PingService
	dht             *dht.IpfsDHT
	discoveryCtx    context.Context
	discoveryCancel context.CancelFunc
	peersLock       sync.RWMutex
	peers           map[peer.ID]peer.AddrInfo
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, port int, bootstrapPeers []string) (*Node, error) {
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

	// Create DHT for peer discovery
	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		h.Close()
		return nil, fmt.Errorf("failed to create DHT: %v", err)
	}

	// Bootstrap the DHT
	if err = kadDHT.Bootstrap(ctx); err != nil {
		h.Close()
		return nil, fmt.Errorf("failed to bootstrap DHT: %v", err)
	}

	// Connect to bootstrap peers if provided
	if len(bootstrapPeers) > 0 {
		for _, addrStr := range bootstrapPeers {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Warn().Err(err).Str("addr", addrStr).Msg("Invalid bootstrap peer address")
				continue
			}

			peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				log.Warn().Err(err).Str("addr", addrStr).Msg("Failed to get peer info from address")
				continue
			}

			if err := h.Connect(ctx, *peerInfo); err != nil {
				log.Warn().Err(err).Str("peer", peerInfo.ID.String()).Msg("Failed to connect to bootstrap peer")
				continue
			}
			log.Info().Str("peer", peerInfo.ID.String()).Msg("Connected to bootstrap peer")
		}
	}

	// Create ping service for node health checks
	ps := ping.NewPingService(h)

	// Create discovery context
	discoveryCtx, discoveryCancel := context.WithCancel(ctx)

	// Create node
	node := &Node{
		Host:            h,
		PingService:     ps,
		dht:             kadDHT,
		discoveryCtx:    discoveryCtx,
		discoveryCancel: discoveryCancel,
		peers:           make(map[peer.ID]peer.AddrInfo),
	}

	// Set up peer discovery
	node.setupDiscovery()

	// Print node info
	fmt.Printf("Node started with ID: %s\n", h.ID().String())
	fmt.Printf("Node addresses:\n")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID().String())
	}

	return node, nil
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

// setupDiscovery configures peer discovery
func (n *Node) setupDiscovery() {
	// Create a routing discovery instance
	routingDiscovery := routing.NewRoutingDiscovery(n.dht)

	// Advertise this node
	routingDiscovery.Advertise(n.discoveryCtx, DiscoveryNamespace)

	// Look for other peers
	go func() {
		for {
			select {
			case <-n.discoveryCtx.Done():
				return
			case <-time.After(DiscoveryInterval):
				log.Info().Msg("Discovering peers...")
				peers, err := routingDiscovery.FindPeers(n.discoveryCtx, DiscoveryNamespace)
				if err != nil {
					fmt.Printf("Failed to find peers: %v\n", err)
					continue
				}

				// Connect to discovered peers
				for peer := range peers {
					log.Info().Msgf("Discovered peer: %s", peer.ID)
					if peer.ID == n.Host.ID() {
						continue // Skip self
					}

					n.peersLock.Lock()
					n.peers[peer.ID] = peer
					n.peersLock.Unlock()

					if err := n.Host.Connect(n.discoveryCtx, peer); err != nil {
						fmt.Printf("Failed to connect to peer %s: %v\n", peer.ID, err)
						continue
					}

					fmt.Printf("Connected to peer: %s\n", peer.ID)
				}
			}
		}
	}()

	// Set up connection handler
	n.Host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			peer := conn.RemotePeer()
			fmt.Printf("Connected to peer: %s\n", peer)
		},
		DisconnectedF: func(net network.Network, conn network.Conn) {
			peer := conn.RemotePeer()
			fmt.Printf("Disconnected from peer: %s\n", peer)
		},
	})
}

// Close shuts down the node
func (n *Node) Close() error {
	// Stop discovery
	n.discoveryCancel()

	// Close DHT
	if err := n.dht.Close(); err != nil {
		fmt.Printf("Error closing DHT: %v\n", err)
	}

	// Close host
	return n.Host.Close()
}

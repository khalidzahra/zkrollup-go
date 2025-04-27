package p2p

import (
	"context"
	"encoding/json"
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

	"zkrollup/pkg/state"
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

	// Protocol handlers
	handlers     *ProtocolHandlers
	handlersLock sync.RWMutex
}

// NewNode creates a new P2P node
func NewNode(ctx context.Context, port int, bootstrapPeers []string) (*Node, error) {
	// Log the node creation
	log.Info().Int("port", port).Msg("Creating new P2P node")
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
		handlers:        &ProtocolHandlers{}, // Initialize empty handlers
	}

	// Register default protocol handlers to ensure basic protocol negotiation works
	node.registerDefaultProtocolHandlers()

	// Set up peer discovery
	node.setupDiscovery()

	// Print node info
	log.Info().Str("id", h.ID().String()).Msg("Node started")
	for _, addr := range h.Addrs() {
		log.Info().Str("addr", addr.Multiaddr().String()+"/p2p/"+h.ID().String()).Msg("Node address")
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
				peers, err := routingDiscovery.FindPeers(n.discoveryCtx, DiscoveryNamespace)
				if err != nil {
					fmt.Printf("Failed to find peers: %v\n", err)
					continue
				}

				// Connect to discovered peers
				for peer := range peers {
					log.Debug().Str("peer", peer.ID.String()).Msg("Discovered peer")
					if peer.ID == n.Host.ID() {
						continue // Skip self
					}

					if n.Host.Network().Connectedness(peer.ID) == network.Connected {
						log.Debug().Str("peer", peer.ID.String()).Msg("Already connected to peer")
						continue // Skip already connected peers
					}

					n.peersLock.Lock()
					n.peers[peer.ID] = peer
					n.peersLock.Unlock()

					if err := n.Host.Connect(n.discoveryCtx, peer); err != nil {
						fmt.Printf("Failed to connect to peer %s: %v\n", peer.ID, err)
						continue
					}

					log.Info().Str("peer", peer.ID.String()).Msg("Connected to peer")
				}
			}
		}
	}()

	// Set up connection handler
	n.Host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			peer := conn.RemotePeer()
			log.Info().Str("peer", peer.String()).Msg("Connected to peer")

			// Log the protocols supported by this peer
			protos, err := n.Host.Peerstore().GetProtocols(peer)
			if err != nil {
				log.Error().Err(err).Str("peer", peer.String()).Msg("Failed to get protocols for peer")
			} else {
				// Convert protocol IDs to strings
				protoStrs := make([]string, len(protos))
				for i, p := range protos {
					protoStrs[i] = string(p)
				}
				log.Info().Strs("protocols", protoStrs).Str("peer", peer.String()).Msg("Peer supports protocols")
			}
		},
		DisconnectedF: func(net network.Network, conn network.Conn) {
			peer := conn.RemotePeer()
			log.Info().Str("peer", peer.String()).Msg("Disconnected from peer")
		},
	})
}

// registerDefaultProtocolHandlers sets up default protocol handlers to ensure they're available for negotiation
func (n *Node) registerDefaultProtocolHandlers() {
	n.handlersLock.RLock()
	handlers := n.handlers
	hasHandlers := handlers != nil &&
		(handlers.OnTransaction != nil || handlers.OnBatch != nil || handlers.OnConsensus != nil)
	n.handlersLock.RUnlock()

	if hasHandlers {
		fmt.Printf("Node %s already has handlers registered, skipping default handlers\n", n.Host.ID().String())
		return
	}

	// Register transaction protocol handler with a no-op handler
	n.Host.SetStreamHandler(TransactionProtocolID, func(s network.Stream) {
		// Check if we have a real handler registered
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		if handlers != nil && handlers.OnTransaction != nil {
			// Use the real handler
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received transaction stream, using registered handler")

			// Read the transaction
			var msg Message
			if err := json.NewDecoder(s).Decode(&msg); err != nil {
				log.Error().Err(err).Msg("Error decoding transaction message")
				s.Reset()
				return
			}

			if msg.Type != MessageTransaction {
				log.Error().Int("type", int(msg.Type)).Msg("Invalid message type for transaction protocol")
				s.Reset()
				return
			}

			var tx state.Transaction
			if err := json.Unmarshal(msg.Payload, &tx); err != nil {
				log.Error().Err(err).Msg("Error unmarshaling transaction")
				s.Reset()
				return
			}

			// Call the transaction handler
			log.Info().Msg("Calling transaction handler")
			if err := handlers.OnTransaction(&tx); err != nil {
				log.Error().Err(err).Msg("Error handling transaction")
				s.Reset()
				return
			}

			log.Info().Msg("Transaction handled successfully")
		} else {
			// No handler registered, just log and close
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received transaction stream with default handler")
		}

		s.Close()
	})

	// Register batch protocol handler with a no-op handler
	n.Host.SetStreamHandler(BatchProtocolID, func(s network.Stream) {
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		if handlers != nil && handlers.OnBatch != nil {
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received batch stream, using registered handler")

			// Read the batch
			var msg Message
			if err := json.NewDecoder(s).Decode(&msg); err != nil {
				log.Error().Err(err).Msg("Error decoding batch message")
				s.Reset()
				return
			}

			if msg.Type != MessageBatch {
				log.Error().Int("type", int(msg.Type)).Msg("Invalid message type for batch protocol")
				s.Reset()
				return
			}

			var batch state.Batch
			if err := json.Unmarshal(msg.Payload, &batch); err != nil {
				log.Error().Err(err).Msg("Error unmarshaling batch")
				s.Reset()
				return
			}

			// Call the batch handler
			log.Info().Msg("Calling batch handler")
			if err := handlers.OnBatch(&batch); err != nil {
				log.Error().Err(err).Msg("Error handling batch")
				s.Reset()
				return
			}

			log.Info().Msg("Batch handled successfully")
		} else {
			// No handler registered, just log and close
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received batch stream with default handler")
		}

		s.Close()
	})

	// Register consensus protocol handler with a no-op handler
	n.Host.SetStreamHandler(ConsensusProtocolID, func(s network.Stream) {
		n.handlersLock.RLock()
		handlers := n.handlers
		n.handlersLock.RUnlock()

		if handlers != nil && handlers.OnConsensus != nil {
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received consensus stream, using registered handler")

			// Read the consensus message
			var msg Message
			if err := json.NewDecoder(s).Decode(&msg); err != nil {
				log.Error().Err(err).Msg("Error decoding consensus message")
				s.Reset()
				return
			}

			if msg.Type != MessageConsensus {
				log.Error().Int("type", int(msg.Type)).Msg("Invalid message type for consensus protocol")
				s.Reset()
				return
			}

			// Call the consensus handler
			log.Info().Msg("Calling consensus handler")
			if err := handlers.OnConsensus(msg.Payload); err != nil {
				log.Error().Err(err).Msg("Error handling consensus message")
				s.Reset()
				return
			}

			log.Info().Msg("Consensus message handled successfully")
		} else {
			// No handler registered, just log and close
			log.Info().Str("peer", s.Conn().RemotePeer().String()).Msg("Received consensus stream with default handler")
		}

		s.Close()
	})

	log.Info().Msg("Registered protocol handlers")
}

// Close shuts down the node
func (n *Node) Close() error {
	// Stop discovery
	n.discoveryCancel()

	// Close DHT
	if err := n.dht.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing DHT")
	}

	// Close host
	return n.Host.Close()
}

// GetProtocolHandlers returns the current protocol handlers
func (n *Node) GetProtocolHandlers() *ProtocolHandlers {
	n.handlersLock.RLock()
	defer n.handlersLock.RUnlock()

	// Return a copy of the handlers to avoid race conditions
	if n.handlers == nil {
		return &ProtocolHandlers{}
	}

	// Create a new handlers struct with the same function references
	return &ProtocolHandlers{
		OnTransaction: n.handlers.OnTransaction,
		OnBatch:       n.handlers.OnBatch,
		OnConsensus:   n.handlers.OnConsensus,
	}
}

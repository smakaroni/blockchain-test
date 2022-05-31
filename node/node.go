package node

import (
	"context"
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"net/http"
)

const (
	DefaultIP      = "127.0.0.1"
	DefaultPort    = 8080
	endpointStatus = "/node/status"

	endpointSync                  = "/node/sync"
	endpointSyncQueryKeyFromBlock = "fromBlock"

	endpointAddPeer             = "/node/peer"
	endpointAddPeerQueryKeyIP   = "ip"
	endpointAddPeerQueryKeyPort = "port"
)

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`

	connected bool
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir    string
	ip         string
	port       uint64
	state      *database.State
	knownPeers map[string]PeerNode
}

func New(dataDir, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:    dataDir,
		ip:         ip,
		port:       port,
		knownPeers: knownPeers,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on: %s:%d", n.ip, n.port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(writer http.ResponseWriter, request *http.Request) {
		listBalancesHandler(writer, request, state)
	})

	http.HandleFunc("/tx/add", func(writer http.ResponseWriter, request *http.Request) {
		txAddHandler(writer, request, state)
	})

	http.HandleFunc(endpointStatus, func(writer http.ResponseWriter, request *http.Request) {
		statusHandler(writer, request, n)
	})

	http.HandleFunc(endpointSync, func(writer http.ResponseWriter, request *http.Request) {
		syncHandler(writer, request, n)
	})

	http.HandleFunc(endpointAddPeer, func(writer http.ResponseWriter, request *http.Request) {
		addPeerHandler(writer, request, n)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

func (n *Node) AddPeer(peer PeerNode) {
	n.knownPeers[peer.TcpAddress()] = peer
}

func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.knownPeers, peer.TcpAddress())
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	if peer.IP == n.ip && peer.Port == n.port {
		return true
	}

	_, isKnownPeer := n.knownPeers[peer.TcpAddress()]

	return isKnownPeer
}

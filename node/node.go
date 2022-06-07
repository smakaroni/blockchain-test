package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"github.com/smakaroni/maaad-blockchain-household/wallet"
	"net/http"
	"time"
)

const (
	DefaultBootstrapIp   = "node.mbh.smakihani"
	DefaultBootstrapPort = 8080
	DefaultBootstrapAcc  = wallet.JokkeAccount
	DefaultMiner         = "0x0000000000000000000000000000000000000000"
	DefaultIP            = "127.0.0.1"
	DefaultPort          = 8080
	endpointStatus       = "/node/status"

	endpointSync                  = "/node/sync"
	endpointSyncQueryKeyFromBlock = "fromBlock"

	endpointAddPeer              = "/node/peer"
	endpointAddPeerQueryKeyIP    = "ip"
	endpointAddPeerQueryKeyPort  = "port"
	endpointAddPeerQueryKeyMiner = "miner"

	miningIntervalSeconds = 10
)

type PeerNode struct {
	IP          string         `json:"ip"`
	Port        uint64         `json:"port"`
	IsBootstrap bool           `json:"is_bootstrap"`
	Account     common.Address `json:"account"`

	connected bool
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir string
	info    PeerNode

	state           *database.State
	knownPeers      map[string]PeerNode
	pendingTXs      map[string]database.SignedTx
	archivedTXs     map[string]database.SignedTx
	newSyncedBlocks chan database.Block
	newPendingTXs   chan database.SignedTx
	isMining        bool
}

func New(dataDir, ip string, port uint64, acc common.Address, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:         dataDir,
		info:            NewPeerNode(ip, port, false, acc, true),
		knownPeers:      knownPeers,
		pendingTXs:      make(map[string]database.SignedTx),
		archivedTXs:     make(map[string]database.SignedTx),
		newSyncedBlocks: make(chan database.Block),
		newPendingTXs:   make(chan database.SignedTx, 10000),
		isMining:        false,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, acc common.Address, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, acc, connected}
}

func (n *Node) Run(ctx context.Context) error {
	fmt.Println(fmt.Sprintf("Listening on %s:%d", n.info.IP, n.info.Port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	fmt.Println("Blockchain state:")
	fmt.Printf("\t -height: %d\n", n.state.LatestBlock().Header.Number)
	fmt.Printf("\t -hashh: %s\n", n.state.LatestBlockHash().Hex())

	go n.sync(ctx)
	go n.mine(ctx)

	http.HandleFunc("/balances/list", func(writer http.ResponseWriter, request *http.Request) {
		listBalancesHandler(writer, request, state)
	})

	http.HandleFunc("/tx/add", func(writer http.ResponseWriter, request *http.Request) {
		txAddHandler(writer, request, n)
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

	server := &http.Server{Addr: fmt.Sprintf(":%d", n.info.Port)}

	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	err = server.ListenAndServe()

	if err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (n *Node) LatestBlockHash() database.Hash {
	return n.state.LatestBlockHash()
}

func (n *Node) mine(ctx context.Context) error {
	var miningCtx context.Context
	var stopCurrentMining context.CancelFunc

	ticker := time.NewTicker(time.Second * miningIntervalSeconds)

	for {
		select {
		case <-ticker.C:
			go func() {
				if len(n.pendingTXs) > 0 && !n.isMining {
					n.isMining = true

					miningCtx, stopCurrentMining = context.WithCancel(ctx)
					err := n.minePendingTXs(miningCtx)
					if err != nil {
						fmt.Errorf("Error %s\n", err)
					}

					n.isMining = false
				}
			}()

		case block, _ := <-n.newSyncedBlocks:
			if n.isMining {
				blockHash, _ := block.Hash()
				fmt.Printf("\nPeer mined next Block %s faster :(\n", blockHash.Hex())

				n.removeMinedPendingTXs(block)
				stopCurrentMining()
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (n *Node) minePendingTXs(ctx context.Context) error {
	blockToMine := NewPendingBlock(
		n.state.LatestBlockHash(),
		n.state.NextBlockNumber(),
		n.info.Account,
		n.getPendingTXsAsArray(),
	)

	minedBlock, err := Mine(ctx, blockToMine)
	if err != nil {
		return err
	}

	n.removeMinedPendingTXs(minedBlock)

	_, err = n.state.AddBlock(minedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) removeMinedPendingTXs(block database.Block) {
	if len(block.TXs) > 0 && len(n.pendingTXs) > 0 {
		fmt.Println("Updating in-memory pending txs pool:")
	}

	for _, tx := range block.TXs {
		txHash, _ := tx.Hash()
		if _, exist := n.pendingTXs[txHash.Hex()]; exist {
			fmt.Printf("\t-archiving mined TX: %s\n", txHash.Hex())

			n.archivedTXs[txHash.Hex()] = tx
			delete(n.pendingTXs, txHash.Hex())
		}
	}
}

func (n *Node) AddPeer(peer PeerNode) {
	n.knownPeers[peer.TcpAddress()] = peer
}

func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.knownPeers, peer.TcpAddress())
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	if peer.IP == n.info.IP && peer.Port == n.info.Port {
		return true
	}

	_, isKnownPeer := n.knownPeers[peer.TcpAddress()]

	return isKnownPeer
}

func (n *Node) AddPendingTX(tx database.SignedTx, fromPeer PeerNode) error {
	txHash, err := tx.Hash()
	if err != nil {
		return err
	}

	txJson, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	_, isAlreadyPending := n.pendingTXs[txHash.Hex()]
	_, isArchived := n.archivedTXs[txHash.Hex()]

	if !isAlreadyPending && !isArchived {
		fmt.Printf("Added Pending TX %s from peer %s\n", txJson, fromPeer.TcpAddress())
		n.pendingTXs[txHash.Hex()] = tx
		n.newPendingTXs <- tx
	}

	return nil
}

func (n *Node) getPendingTXsAsArray() []database.SignedTx {
	txs := make([]database.SignedTx, len(n.pendingTXs))

	i := 0
	for _, tx := range n.pendingTXs {
		txs[i] = tx
		i++
	}

	return txs
}

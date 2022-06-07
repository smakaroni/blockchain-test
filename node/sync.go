package node

import (
	"context"
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"net/http"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-ticker.C:
			n.doSync()

		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) doSync() {
	for _, peer := range n.knownPeers {
		if n.info.IP == peer.IP && n.info.Port == peer.Port {
			continue
		}

		fmt.Printf("Searching for new Peers and their Blocks and peers: %s\n", peer.TcpAddress())

		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			fmt.Printf("Peer %s was removed from known peer \n", peer.TcpAddress())

			n.RemovePeer(peer)

			continue
		}

		err = n.joinKnownPeers(peer)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		err = n.syncBlocks(peer, status)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		err = n.syncKnownPeers(peer, status)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		err = n.syncPendingTXs(peer, status.PendingTXs)
		if err != nil {
			fmt.Printf("Error %s\n", err.Error())
			continue
		}
	}
}

func (n *Node) syncBlocks(peer PeerNode, status StatusRes) error {
	localBlockNumber := n.state.LatestBlock().Header.Number

	//ignore peer with no blocks
	if status.Hash.IsEmpty() {
		return nil
	}

	//ignore if peer has less blocks than us
	if status.Number < localBlockNumber {
		return nil
	}

	// ignore genesis block and already synced
	if status.Number == 0 && !n.state.LatestBlockHash().IsEmpty() {
		return nil
	}

	//display found 1 new block if we sync the genesis block 0
	newBlockCount := status.Number - localBlockNumber
	if localBlockNumber == 0 && status.Number == 0 {
		newBlockCount = 1
	}
	fmt.Printf("Found %d new blocks from Peer %s\n", newBlockCount, peer.TcpAddress())

	blocks, err := fetchBlockFromPeer(peer, n.state.LatestBlockHash())
	if err != nil {
		return err
	}

	for _, block := range blocks {
		_, err = n.state.AddBlock(block)
		if err != nil {
			return err
		}

		n.newSyncedBlocks <- block
	}

	return nil
}

func (n *Node) syncKnownPeers(peer PeerNode, status StatusRes) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(statusPeer) {
			fmt.Printf("Found new Peer %s\n", statusPeer.TcpAddress())

			n.AddPeer(statusPeer)
		}
	}

	return nil
}

func (n *Node) syncPendingTXs(peer PeerNode, txs []database.SignedTx) error {
	for _, tx := range txs {
		err := n.AddPendingTX(tx, peer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) joinKnownPeers(peer PeerNode) error {
	if peer.connected {
		return nil
	}

	url := fmt.Sprintf(
		"http://%s%s?%s=%s&%s=%d",
		peer.TcpAddress(),
		endpointAddPeer,
		endpointAddPeerQueryKeyIP,
		n.info.IP,
		endpointAddPeerQueryKeyPort,
		n.info.Port,
	)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	addPeerRes := AddPeerRes{}
	err = readRes(res, &addPeerRes)
	if err != nil {
		return err
	}
	if addPeerRes.Error != "" {
		return fmt.Errorf(addPeerRes.Error)
	}

	knownPeer := n.knownPeers[peer.TcpAddress()]
	knownPeer.connected = addPeerRes.Success

	n.AddPeer(knownPeer)

	if !addPeerRes.Success {
		return fmt.Errorf("unable to join known peer os %s", peer.TcpAddress())
	}

	return nil
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s%s", peer.TcpAddress(), endpointStatus)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, err
	}

	statusRes := StatusRes{}
	err = readRes(res, &statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}

func fetchBlockFromPeer(peer PeerNode, fromBlock database.Hash) ([]database.Block, error) {
	fmt.Printf("Importing blocks from Peer %s...\n", peer.TcpAddress())

	url := fmt.Sprintf(
		"http://%s%s?%s=%s",
		peer.TcpAddress(),
		endpointSync,
		endpointSyncQueryKeyFromBlock,
		fromBlock.Hex(),
	)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	syncRes := SyncRes{}
	err = readRes(res, &syncRes)
	if err != nil {
		return nil, err
	}

	return syncRes.Blocks, nil
}

package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"reflect"
	"sort"
)

const TxFee = uint(50)

type State struct {
	Balances      map[common.Address]uint
	Account2Nonce map[common.Address]uint

	dbFile *os.File

	latestBlock     Block
	latestBlockHash Hash
	hasGenesisBlock bool

	difficulty uint
}

func NewStateFromDisk(dataDir string, difficulty uint) (*State, error) {
	err := initDataDirIfNotExist(dataDir, []byte(genesisJson))
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonPath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[common.Address]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	account2nonce := make(map[common.Address]uint)

	f, err := os.OpenFile(getBlocksDbFile(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{balances, account2nonce, f, Block{}, Hash{}, false, difficulty}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()

		if len(blockFsJson) == 0 {
			break
		}

		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		err = applyBlock(blockFs.Value, state)
		if err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFs.Key
		state.latestBlock = blockFs.Value
		state.hasGenesisBlock = true
	}

	return state, nil
}

func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := s.AddBlock(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	err := applyBlock(b, &pendingState)
	if err != nil {
		return Hash{}, err
	}

	blockhash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{blockhash, b}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisiting new Bloc to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	_, err = s.dbFile.Write(append(blockFsJson, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = pendingState.Balances
	s.Account2Nonce = pendingState.Account2Nonce
	s.latestBlockHash = blockhash
	s.latestBlock = b
	s.hasGenesisBlock = true

	return blockhash, nil
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LatestBlock().Header.Number + 1
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) GetNextAccountNonce(account common.Address) uint {
	return s.Account2Nonce[account] + 1
}

func (s *State) Close() error {
	return s.dbFile.Close()
}

func (s *State) copy() State {
	c := State{}
	c.hasGenesisBlock = s.hasGenesisBlock
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.Balances = make(map[common.Address]uint)
	c.Account2Nonce = make(map[common.Address]uint)

	for acc, bal := range s.Balances {
		c.Balances[acc] = bal
	}

	for acc, nonce := range s.Account2Nonce {
		c.Account2Nonce[acc] = nonce
	}

	return c
}

func applyBlock(b Block, s *State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	hash, err := b.Hash()
	if err != nil {
		return err
	}

	if !IsBlockHashValid(hash, s.difficulty) {
		return fmt.Errorf("invalid block hash %x", hash)
	}

	err = applyTxs(b.TXs, s)
	if err != nil {
		return err
	}

	s.Balances[b.Header.Miner] += BlockReward
	s.Balances[b.Header.Miner] += uint(len(b.TXs)) * TxFee

	return nil
}

func applyTxs(txs []SignedTx, s *State) error {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Time < txs[j].Time
	})

	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyTx(tx SignedTx, s *State) error {
	err := ValidateTx(tx, s)
	if err != nil {
		return err
	}

	s.Balances[tx.From] -= tx.Cost()
	s.Balances[tx.To] += tx.Value
	s.Account2Nonce[tx.From] = tx.Nonce

	return nil
}

func ValidateTx(tx SignedTx, s *State) error {
	ok, err := tx.IsAuthentic()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("wrong TX, Sender %s is forged", tx.From.String())
	}

	expectedNonce := s.GetNextAccountNonce(tx.From)
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("wrong tx sender %s next nonce must be %d not %d", tx.From.String(), expectedNonce, tx.Nonce)
	}

	if tx.Cost() > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender %s balance is %d MBH. Tx cost is %d MBH", tx.From.String(), s.Balances[tx.From], tx.Cost())
	}

	return nil
}

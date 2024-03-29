package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
)

var genesisJson = `
{
  "genesis_time": "2022-05-27T00:00.000000000Z",
  "chain_id": "the-maaad-blockchain-household",
  "balances": {
    "0xEdD144f5D916340F285d5F34309B4E8b65A65570": 1000000
  }
}`

type genesis struct {
	Balances map[common.Address]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGen genesis
	err = json.Unmarshal(content, &loadedGen)
	if err != nil {
		return genesis{}, err
	}

	return loadedGen, nil
}

func writeGenesisToDisk(path string, genesis []byte) error {
	return ioutil.WriteFile(path, genesis, 0644)
}

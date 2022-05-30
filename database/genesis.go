package database

import (
	"encoding/json"
	"io/ioutil"
)

var genesisJson = `
{
  "genesis_time": "2022-05-27T00:00.000000000Z",
  "chain_id": "the-maaad-blockchain-household",
  "balances": {
    "jokke": 1000000
  }
}`

type genesis struct {
	Balances map[Account]uint `json:"balances"`
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

func writeGenesisToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(genesisJson), 0644)
}

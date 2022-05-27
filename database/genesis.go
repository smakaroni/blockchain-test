package database

import (
	"encoding/json"
	"io/ioutil"
)

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

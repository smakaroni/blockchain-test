package database

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
)

func GetBlocksAfter(blockHash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(getBlocksDbFile(dataDir), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	shouldStartCollectin := false

	if reflect.DeepEqual(blockHash, Hash{}) {
		shouldStartCollectin = true
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFs BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blockFs)
		if err != nil {
			return nil, err
		}

		if shouldStartCollectin {
			blocks = append(blocks, blockFs.Value)
			continue
		}

		if blockHash == blockFs.Key {
			shouldStartCollectin = true
		}
	}

	return blocks, nil
}

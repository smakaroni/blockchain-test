package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func initDataDirIfNotExist(dataDir string, genesis []byte) error {
	if fileExist(getGenesisJsonPath(dataDir)) {
		return nil
	}

	if err := os.MkdirAll(getDbDirPath(dataDir), os.ModePerm); err != nil {
		return err
	}

	if err := writeGenesisToDisk(getGenesisJsonPath(dataDir), genesis); err != nil {
		return err
	}

	if err := writeEmptyBlocksDbToDisk(getBlocksDbFile(dataDir)); err != nil {
		return err
	}

	return nil
}

func getDbDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonPath(dataDir string) string {
	return filepath.Join(getDbDirPath(dataDir), "genesis.json")
}

func getBlocksDbFile(dataDir string) string {
	return filepath.Join(getDbDirPath(dataDir), "block.db")
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func writeEmptyBlocksDbToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}

package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"os"
	"time"
)

func main() {
	cwd, _ := os.Getwd()
	state, err := database.NewStateFromDisk(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("jokke", "emma", 500000, "hej"),
			database.NewTx("emma", "jokke", 10, "hej"),
		},
	)

	state.AddBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("jokke", "emma", 10, ""),
			database.NewTx("jokke", "majken", 1000, ""),
			database.NewTx("emma", "majken", 1000, ""),
		},
	)

	state.AddBlock(block1)
	state.Persist()
}

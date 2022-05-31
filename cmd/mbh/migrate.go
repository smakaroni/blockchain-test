package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain db",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := database.NewBlock(
				database.Hash{},
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("jokke", "emma", 500000, ""),
				})

			block0Hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(
				block0Hash,
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("emma", "majken", 1000, ""),
					database.NewTx("jokke", "majken", 1000, ""),
				})

			_, err = state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultReqFlags(migrateCmd)

	return migrateCmd
}

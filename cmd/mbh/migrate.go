package main

import (
	"context"
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"github.com/smakaroni/maaad-blockchain-household/node"
	"github.com/spf13/cobra"
	"time"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain db",
		Run: func(cmd *cobra.Command, args []string) {
			miner, _ := cmd.Flags().GetString(flagMiner)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)

			peer := node.NewPeerNode("127.0.0.1", 8080, true, database.NewAccount("jokke"), false)

			n := node.New(getDataDirFromCmd(cmd), ip, port, database.NewAccount(miner), peer)

			n.AddPendingTX(database.NewTx("jokke", "emma", 500000, ""), peer)
			n.AddPendingTX(database.NewTx("jokke", "majken", 1000, ""), peer)
			n.AddPendingTX(database.NewTx("emma", "majken", 1000, ""), peer)

			ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*15)

			go func() {
				ticker := time.NewTicker(time.Second * 10)

				for {
					select {
					case <-ticker.C:
						if !n.LatestBlockHash().IsEmpty() {
							closeNode()
							return
						}
					}
				}
			}()

			err := n.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	addDefaultReqFlags(migrateCmd)
	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards")
	migrateCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultPort, "exposed HTTP port for communication with peers")

	return migrateCmd
}

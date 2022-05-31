package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/node"
	"github.com/spf13/cobra"
	"os"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the MBH node and its API",
		Run: func(cmd *cobra.Command, args []string) {
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)

			fmt.Println("Launching MBH...")

			bootStrap := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				false)

			n := node.New(getDataDirFromCmd(cmd), ip, port, bootStrap)
			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultReqFlags(runCmd)
	runCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	runCmd.Flags().Uint64(flagPort, node.DefaultPort, "exposed port for communication with peers")
	return runCmd
}

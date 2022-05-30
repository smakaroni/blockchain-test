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
			dataDir, _ := cmd.Flags().GetString(flagDataDir)

			fmt.Println("Launching MBH...")

			err := node.Run(dataDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultReqFlags(runCmd)
	return runCmd
}

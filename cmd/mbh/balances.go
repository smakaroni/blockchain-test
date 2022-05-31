package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"github.com/spf13/cobra"
	"os"
)

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances (list...).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	balancesCmd.AddCommand(balanceListCmd())

	return balancesCmd
}

func balanceListCmd() *cobra.Command {
	var balanceListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all balances",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			fmt.Printf("Accounts balances at %x:\n", state.LatestBlockHash())
			fmt.Println("_________________________")
			fmt.Println("")
			for account, balance := range state.Balances {
				fmt.Println(fmt.Sprintf("%s: %d", account, balance))
			}
		},
	}

	addDefaultReqFlags(balanceListCmd)

	return balanceListCmd
}

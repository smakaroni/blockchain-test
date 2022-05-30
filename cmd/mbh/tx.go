package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/database"
	"github.com/spf13/cobra"
	"os"
)

const (
	flagFrom  = "from"
	flagTo    = "to"
	flagValue = "value"
	flagData  = "data"
)

func txCmd() *cobra.Command {
	var transactionsCmd = &cobra.Command{
		Use:   "tx",
		Short: "Interact with transaction",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	transactionsCmd.AddCommand(transactionAddCmd())

	return transactionsCmd
}

func transactionAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Adds new transaction to db",
		Run: func(cmd *cobra.Command, args []string) {
			//TODO error hading
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _ := cmd.Flags().GetUint(flagValue)
			data, _ := cmd.Flags().GetString(flagData)

			tx := database.NewTx(database.NewAccount(from), database.NewAccount(to), value, data)

			state, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			err = state.AddTx(tx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			_, err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("Transaction added to the ledger")
		},
	}

	cmd.Flags().String(flagFrom, "", "From what account to send tokens")
	cmd.MarkFlagRequired(flagFrom)

	cmd.Flags().String(flagTo, "", "To what account to send tokens")
	cmd.MarkFlagRequired(flagTo)

	cmd.Flags().Uint(flagValue, 0, "How many tokens to send")
	cmd.MarkFlagRequired(flagValue)

	cmd.Flags().String(flagData, "", "Possible values: 'reward'")

	return cmd
}

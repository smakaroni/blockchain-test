package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var mbhCmd = &cobra.Command{
		Use:   "mbh",
		Short: "The Maaad Blockchain Household CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	mbhCmd.AddCommand(versionCmd)
	mbhCmd.AddCommand(balancesCmd())
	mbhCmd.AddCommand(txCmd())

	err := mbhCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

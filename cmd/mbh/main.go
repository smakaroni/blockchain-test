package main

import (
	"fmt"
	"github.com/smakaroni/maaad-blockchain-household/fs"
	"github.com/spf13/cobra"
	"os"
)

const (
	flagDataDir = "datadir"
	flagIP      = "ip"
	flagPort    = "port"
)

func main() {
	var mbhCmd = &cobra.Command{
		Use:   "mbh",
		Short: "The Maaad Blockchain Household CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	mbhCmd.AddCommand(migrateCmd())
	mbhCmd.AddCommand(versionCmd)
	mbhCmd.AddCommand(runCmd())
	mbhCmd.AddCommand(balancesCmd())

	err := mbhCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultReqFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dir where the DB is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)
	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

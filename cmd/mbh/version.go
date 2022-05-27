package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	Major  = "0"
	Minor  = "1"
	Fix    = "0"
	Verbal = "TX Add && Balances List"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal))
	},
}

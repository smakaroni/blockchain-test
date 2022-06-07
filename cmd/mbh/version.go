package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	Major  = "1"
	Minor  = "0"
	Fix    = "0"
	Verbal = "Decentralized authentication"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal))
	},
}

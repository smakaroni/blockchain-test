package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	Major  = "0"
	Minor  = "7"
	Fix    = "2"
	Verbal = "Sync"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal))
	},
}

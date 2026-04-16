package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of sd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sd CLI v%s\n", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("1.2.2")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

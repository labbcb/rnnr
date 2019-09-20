package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var database, address, host string
var cpuCores int
var ramGb float64

var rootCmd = &cobra.Command{
	Use:   "rnnr",
	Short: "Distributed task executor for genomic research",
}

// Execute calls root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

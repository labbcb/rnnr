package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
)

func init() {
	viewCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8080", "URL to RNNR server")
	rootCmd.AddCommand(viewCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		t, err := client.GetTask(host, args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to get task:", err)
			os.Exit(1)
		}
		json.NewEncoder(os.Stdout).Encode(t)
	},
}

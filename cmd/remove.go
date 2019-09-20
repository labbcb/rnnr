package cmd

import (
	"github.com/spf13/cobra"
	"github.com/labbcb/rnnr/client"
	"log"
)

func init() {
	removeCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8080", "URL to RNNR server")
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove NODE_URL",
	Short: "Remove a computing note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		if err := client.DisableNode(host, id); err != nil {
			log.Fatalf("Unable to deactivate worker %s: %s\n", id, err)
		}
	},
}

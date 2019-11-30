package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"log"
)

var disableCmd = &cobra.Command{
	Use:   "remove url...",
	Short: "Remove one or more computing nodes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			if err := client.DisableNode(host, id); err != nil {
				log.Printf("Unable to deactivate worker %s: %s\n", id, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}

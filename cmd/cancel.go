package cmd

import (
	"github.com/spf13/cobra"
	"github.com/labbcb/rnnr/client"
	"log"
)

func init() {
	cancelCmd.Flags().StringVar(&host, "host", "http://localhost:8080", "URL to RNNR server")
	rootCmd.AddCommand(cancelCmd)
}

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		if err := client.CancelTask(host, id); err != nil {
			log.Fatalf("Unable to cancel task %s: %v\n", id, err)
		}
	},
}

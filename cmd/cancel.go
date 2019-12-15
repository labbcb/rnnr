package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var cancelCmd = &cobra.Command{
	Use:     "cancel id...",
	Aliases: []string{"abort", "stop"},
	Short:   "Cancel one or more tasks",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		for _, id := range args {
			if err := client.CancelTask(host, id); err != nil {
				log.Printf("Unable to cancel models %s: %v\n", id, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}

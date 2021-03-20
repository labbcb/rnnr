package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/viper"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:     "view id...",
	Aliases: []string{"get"},
	Short:   "Get one or more tasks by their IDs",
	Long:    "Use --format json to print in JSON format.",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		for _, id := range args {
			task, err := client.GetTask(host, id, models.Full)
			if err != nil {
				message("%v\n", err)
				continue
			}

			if err := json.NewEncoder(os.Stdout).Encode(task); err != nil {
				message("%v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

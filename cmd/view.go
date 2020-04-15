package cmd

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:     "view id...",
	Aliases: []string{"get"},
	Short:   "View one or more tasks",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		for _, id := range args {
			task, err := client.GetTask(host, id)
			fatalOnErr(err)

			if err := json.NewEncoder(os.Stdout).Encode(task); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

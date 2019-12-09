package cmd

import (
	"encoding/json"
	"github.com/spf13/viper"
	"log"
	"os"

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
			t, err := client.GetTask(host, id)
			if err != nil {
				log.Println(err)
			} else {
				if err := json.NewEncoder(os.Stdout).Encode(t); err != nil {
					log.Println(err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

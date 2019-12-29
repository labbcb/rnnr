package cmd

import (
	"github.com/labbcb/rnnr/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var disableCmd = &cobra.Command{
	Use:     "disable url...",
	Aliases: []string{"remove", "rm"},
	Short:   "Remove one or more computing nodes",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
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

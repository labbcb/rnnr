package cmd

import (
	"github.com/labbcb/rnnr/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cancel bool

var disableCmd = &cobra.Command{
	Use:     "disable url...",
	Aliases: []string{"remove", "rm"},
	Short:   "Disable or more worker nodes. Tasks will keep running at disabled node but no tasks will be delagated to worker. Cancel option tells server to cancel tasks in node and enqueue those tasks.",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		for _, id := range args {
			if err := client.DisableNode(host, id, cancel); err != nil {
				log.Printf("Unable to disable worker %s: %s\n", id, err)
			}
		}
	},
}

func init() {
	disableCmd.Flags().BoolVarP(&cancel, "cancel", "c", false, "Cancel running tasks in worker node.")
	rootCmd.AddCommand(disableCmd)
}

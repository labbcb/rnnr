package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var allTasks bool

var cancelCmd = &cobra.Command{
	Use:     "cancel [--all] [id]...",
	Aliases: []string{"abort", "stop"},
	Short:   "Stop one or more tasks",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")

		if allTasks {
			activeStates := []models.State{models.Queued, models.Initializing, models.Running, models.Paused}
			tasks, err := client.ListTasks(host, 0, "", models.Minimal, activeStates)
			fatalOnErr(err)
			for _, task := range tasks.Tasks {
				if err := client.CancelTask(host, task.ID); err != nil {
					log.Printf("Unable to cancel models %s: %v\n", task.ID, err)
					continue
				}
			}
		}

		for _, id := range args {
			if err := client.CancelTask(host, id); err != nil {
				log.Printf("Unable to cancel models %s: %v\n", id, err)
			}
		}
	},
}

func init() {
	cancelCmd.Flags().BoolVarP(&allTasks, "all", "a", false, "Cancel all active tasks.")
	rootCmd.AddCommand(cancelCmd)
}

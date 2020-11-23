package cmd

import (
	"fmt"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var allTasks bool

var cancelCmd = &cobra.Command{
	Use:     "cancel [--all] [id]...",
	Aliases: []string{"abort", "stop"},
	Short:   "Cancel one or more active tasks",
	Long: "This command tells main server to stop active tasks.\n" +
		"--all flag will cancel all active tasks!\n" +
		"It will print IDs of successfully cancelled tasks.",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")

		if all && len(args) > 0 {
			messageAndExit("Use either --all or tasks IDs but not both.")
		}

		if allTasks {
			tasks, err := client.ListTasks(host, 0, "", models.Minimal, nil, models.ActiveStates())
			exitOnErr(err)
			for _, task := range tasks.Tasks {
				if err := client.CancelTask(host, task.ID); err != nil {
					message("Unable to cancel task %s: %v\n", task.ID, err)
					continue
				}
				fmt.Println(task.ID)
			}
		}

		for _, id := range args {
			if err := client.CancelTask(host, id); err != nil {
				message("Unable to cancel task %s: %v\n", id, err)
				continue
			}
			fmt.Println(id)
		}
	},
}

func init() {
	cancelCmd.Flags().BoolVarP(&allTasks, "all", "a", false, "Cancel all active tasks.")
	rootCmd.AddCommand(cancelCmd)
}

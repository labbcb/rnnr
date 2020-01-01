package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

var all bool

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "ListTasks tasks",
	Long:    "It will print only QUEUED and RUNNING tasks by default.",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.ListTasks(host)
		fatalOnErr(err)

		fmt.Printf("%-36s   %-18s   %-14s   %s\n", "Task ID", "Resources", "State", "Name at server (elapsed time)")

		var desc string
		for _, task := range resp.Tasks {
			if !(all || task.Active()) {
				continue
			}
			if task.RemoteHost != "" {
				desc = task.Name + " at " + task.RemoteHost
			} else {
				desc = task.Name
			}

			switch task.State {
			case models.Queued:
			case models.Running:
				desc = fmt.Sprintf("%s (%s)", desc, time.Since(task.Logs.StartTime))
			default:
				desc = fmt.Sprintf("%s (%s)", desc, task.Logs.EndTime.Sub(task.Logs.StartTime))
			}
			fmt.Printf("%36s | CPU=%02d RAM=%05.2fGB | %-14s | %s\n",
				task.ID, task.Resources.CPUCores, task.Resources.RAMGb, task.State, desc)
		}
	},
}

func init() {
	tasksCmd.Flags().BoolVarP(&all, "all", "a", false, "Print all tasks.")
	rootCmd.AddCommand(tasksCmd)
}

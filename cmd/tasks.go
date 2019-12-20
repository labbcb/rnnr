package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var all bool

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List tasks",
	Long:    "It will print only QUEUED and RUNNING tasks by default.",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.ListTasks(host)
		fatalOnErr(err)
		fmt.Printf("%-36s   %-18s   %-14s   %s\n", "Task ID", "Resources", "State", "Name")

		var name string
		for _, t := range resp.Tasks {
			if !(all || t.State == models.Queued || t.State == models.Running) {
				continue
			}
			if t.RemoteHost != "" {
				name = t.Name + " at " + t.RemoteHost
			} else {
				name = t.Name
			}
			fmt.Printf("%36s | CPU=%02d RAM=%05.2fGB | %-14s | %s\n",
				t.ID, t.Resources.CPUCores, t.Resources.RAMGb, t.State, name)
		}
	},
}

func init() {
	tasksCmd.Flags().BoolVarP(&all, "all", "a", false, "Print all tasks.")
	rootCmd.AddCommand(tasksCmd)
}

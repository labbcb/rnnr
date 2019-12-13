package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.ListTasks(host)
		fatalOnErr(err)
		fmt.Printf("%-36s  %-18s   %-14s   %s\n", "Task ID", "Resources", "State", "Name (Server)")

		for _, t := range resp.Tasks {
			fmt.Printf("%36s | CPU=%02d RAM=%05.2fGB | %-14s | %s (%s)\n",
				t.ID, t.Resources.CPUCores, t.Resources.RAMGb, t.State, t.Name, t.RemoteHost)
		}
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}

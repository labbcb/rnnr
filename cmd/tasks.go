package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.ListTasks(host)
		fatalOnErr(err)
		for _, t := range resp.Tasks {
			fmt.Printf("%s CPU=%02d RAM=%05.2fGB %-13s %s\n",
				t.ID, t.Resources.CPUCores, t.Resources.RAMGb, t.State, t.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}

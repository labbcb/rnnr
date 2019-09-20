package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/labbcb/rnnr/client"
)

func init() {
	tasksCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8080", "URL to RNNR server")
	rootCmd.AddCommand(tasksCmd)
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := client.ListTasks(host)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to list tasks:", err)
			os.Exit(1)
		}
		for _, t := range resp.Tasks {
			fmt.Printf("%s CPU=%02d RAM=%05.2fGB %-13s %s\n",
				t.ID, t.Resources.CPUCores, t.Resources.RAMGb, t.State, t.Name)
		}
	},
}

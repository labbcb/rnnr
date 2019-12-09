package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List tasks",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.ListTasks(host)
		fatalOnErr(err)
		fmt.Printf("%-36s   %-19s   %-19s   %-19s   %-18s   %-14s   %s\n",
			"Task ID", "Created", "Started", "Completed", "Resources", "State", "Name")

		var start, end string

		for _, t := range resp.Tasks {
			if t.Logs != nil {
				start = formatDateTime(t.Logs.StartTime)
				end = formatDateTime(t.Logs.EndTime)
			} else {
				start = ""
				end = ""
			}

			fmt.Printf("%36s | %-19s | %-19s | %-19s | CPU=%02d RAM=%05.2fGB | %-14s | %s\n",
				t.ID,
				formatDateTime(t.CreationTime),
				start,
				end,
				t.Resources.CPUCores,
				t.Resources.RAMGb,
				t.State,
				t.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}

func formatDateTime(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

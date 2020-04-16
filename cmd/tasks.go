package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var all bool

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List active tasks",
	Long:    "It will print only active tasks (QUEUED and RUNNING) tasks by default.",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.ListTasks(host)
		fatalOnErr(err)

		format := viper.GetString("format")
		if format == "json" {
			if err := json.NewEncoder(os.Stdout).Encode(resp.Tasks); err != nil {
				log.Fatal(err)
			}
			return
		}

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

			fmt.Printf("%36s | CPU=%02d RAM=%05.2fGB | %-14s | %s (%s)\n",
				task.ID, task.Resources.CPUCores, task.Resources.RAMGb, task.State, desc, task.Elapsed())
		}
	},
}

func init() {
	tasksCmd.Flags().BoolVarP(&all, "all", "a", false, "Print all tasks.")
	rootCmd.AddCommand(tasksCmd)
}

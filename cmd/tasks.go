package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var all, errors bool

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "List active tasks",
	Long:    "It will print only active tasks (QUEUED and RUNNING) tasks by default.",
	Run: func(cmd *cobra.Command, args []string) {
		if all && errors {
			log.Fatal("Use either --all or --error flags but not both.")
		}

		var states []models.State
		if errors {
			states = []models.State{models.ExecutorError, models.SystemError}
		} else if !all {
			states = []models.State{models.Queued, models.Initializing, models.Running, models.Paused}
		}

		host := viper.GetString("host")
		resp, err := client.ListTasks(host, 0, "", models.Basic, states)
		fatalOnErr(err)

		format := viper.GetString("format")
		if format == "json" {
			err := json.NewEncoder(os.Stdout).Encode(resp.Tasks)
			fatalOnErr(err)
		}

		var line string
		for _, task := range resp.Tasks {
			line = fmt.Sprintf("%36s | CPU=%02d RAM=%05.2fGB", task.ID, task.Resources.CPUCores, task.Resources.RAMGb)

			if all || errors {
				line = fmt.Sprintf("%s | %-14s", line, task.State)
			} else {
				line = fmt.Sprintf("%s | %-8s", line, task.State)
			}

			if task.RemoteHost != "" {
				line = fmt.Sprintf("%s | %s at %s (%s)", line, task.Name, task.RemoteHost, task.Elapsed())
			} else {
				line = fmt.Sprintf("%s | %s", line, task.Name)
			}

			fmt.Println(line)
		}
	},
}

func init() {
	tasksCmd.Flags().BoolVarP(&all, "all", "a", false, "Print all tasks.")
	tasksCmd.Flags().BoolVarP(&errors, "error", "e", false, "Print only tasks with error states.")
	rootCmd.AddCommand(tasksCmd)
}

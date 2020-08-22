package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var all, errors bool
var nodes, states []string

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"ls", "list"},
	Short:   "Get tasks",
	Long: "It will print only active tasks by default.\n" +
		"Use --all to get all tasks or --error to get failed tasks.\n" +
		"Use one or more --state parameter to filter by task states.\n" +
		"Valid states are queued, initializing, running, paused, complete,\n" +
		"executor_error, system_error and canceled. States are case insensitive.\n" +
		"Use one or more --node parameter to filter by worker nodes.\n" +
		"Use --format csv to export as CSV.",
	Run: func(cmd *cobra.Command, args []string) {
		if all && errors {
			messageAndExit("Use either --all or --error flags but not both.")
		}
		if len(states) > 0 && (all || errors) {
			messageAndExit("Use either --all or --error flags or --state parameter.")
		}

		var filterStates []models.State
		if errors {
			filterStates = models.ErrorStates()
		} else if !all {
			filterStates = models.ActiveStates()
		} else if len(states) > 0 {
			for _, state := range states {
				filterStates = append(filterStates, models.State(strings.ToUpper(state)))
			}
		}

		host := viper.GetString("host")
		resp, err := client.ListTasks(host, 0, "", models.Basic, nodes, filterStates)
		exitOnErr(err)

		format := viper.GetString("format")
		if format == "json" {
			err := json.NewEncoder(os.Stdout).Encode(resp.Tasks)
			exitOnErr(err)
			return
		}

		if format == "csv" {
			w := csv.NewWriter(os.Stdout)
			w.Write([]string{
				"id",
				"name",
				"description",
				"state",
				"created",
				"cpu_cores",
				"memory_gb",
				"worker_node",
				"cpu_time",
				"max_cpu_percentage",
				"max_memory_bytes",
				"started",
				"completed",
				"executor_started",
				"executor_completed",
				"exit_code",
			})

			for _, t := range resp.Tasks {
				var started, completed, executorStarted, executorCompleted, exitCode string
				if t.Terminated() {
					started = t.Logs[0].StartTime.String()
					completed = t.Logs[0].EndTime.String()
					executorStarted = t.Logs[0].ExecutorLogs[0].StartTime.String()
					executorCompleted = t.Logs[0].ExecutorLogs[0].EndTime.String()
					exitCode = strconv.FormatInt(int64(t.Logs[0].ExecutorLogs[0].ExitCode), 10)
				}

				r := []string{
					t.ID,
					t.Name,
					t.Description,
					string(t.State),
					t.Created.String(),
					strconv.FormatInt(int64(t.Resources.CPUCores), 10),
					strconv.FormatFloat(t.Resources.RAMGb, 'f', -1, 64),
					t.Host,
					strconv.FormatUint(t.Metrics.CPUTime, 10),
					strconv.FormatFloat(t.Metrics.CPUPercentage, 'f', -1, 64),
					strconv.FormatUint(t.Metrics.Memory, 10),
					started,
					completed,
					executorStarted,
					executorCompleted,
					exitCode,
				}
				err := w.Write(r)
				exitOnErr(err)
			}

			w.Flush()
			exitOnErr(w.Error())
			return
		}

		var line string
		for _, task := range resp.Tasks {
			line = fmt.Sprintf("%36s | CPU=%02d RAM=%05.2fGB", task.ID, task.Resources.CPUCores, task.Resources.RAMGb)

			if all || errors {
				line = fmt.Sprintf("%s | %-14s", line, task.State)
			} else {
				line = fmt.Sprintf("%s | %-8s", line, task.State)
			}

			if task.Host != "" {
				line = fmt.Sprintf("%s | %s at %s (%s)", line, task.Name, task.Host, task.Elapsed())
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
	tasksCmd.Flags().StringArrayVarP(&nodes, "node", "n", nil, "Filter tasks by worker nodes.")
	tasksCmd.Flags().StringArrayVarP(&states, "state", "s", nil, "Filter tasks by task states.")
	rootCmd.AddCommand(tasksCmd)
}

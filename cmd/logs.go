package cmd

import (
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
)

var stdout, stderr bool

func init() {
	logsCmd.Flags().StringVar(&host, "host", "http://localhost:8080", "URL to RNNR server")
	logsCmd.Flags().BoolVar(&stdout, "stdout", false, "Prints Task.Executor standard out")
	logsCmd.Flags().BoolVar(&stderr, "stderr", false, "Prints Task.Executor standard error")
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Task logs",
	Long: `Task provides many logs. By default it prints Task.Log.SystemLogs.
	Use --stdout and --stderr to get Task.Log.ExecutorLogs.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		t, err := client.GetTask(host, args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to list tasks:", err)
			os.Exit(1)
		}

		if t.Active() {
			return
		}
		if stdout || stderr {
			if stdout {
				println(t.Logs.Logs[0].Stdout)
			}
			if stderr {
				println(t.Logs.Logs[0].Stderr)
			}
		} else {
			println(t.Logs.SystemLogs)
		}
	},
}

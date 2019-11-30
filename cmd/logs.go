package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"log"
)

var stdout, stderr bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Task logs",
	Long: `Task provides many logs. By default it prints Task.Log.SystemLogs.
	Use --stdout and --stderr to get Task.Log.ExecutorLogs.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		t, err := client.GetTask(host, args[0])
		if err != nil {
			log.Fatalf("Unable to list tasks: %v", err)
		}

		if t.Active() {
			return
		}
		if stdout || stderr {
			if stdout {
				println(t.Logs.ExecutorLogs[0].Stdout)
			}
			if stderr {
				println(t.Logs.ExecutorLogs[0].Stderr)
			}
		} else {
			println(t.Logs.SystemLogs)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVar(&stdout, "stdout", false, "Prints Task.Executor standard out")
	logsCmd.Flags().BoolVar(&stderr, "stderr", false, "Prints Task.Executor standard error")
	rootCmd.AddCommand(logsCmd)
}

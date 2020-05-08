package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var stdout, stderr bool

var logsCmd = &cobra.Command{
	Use:   "logs id",
	Short: "Get task logs",
	Long: "Task provides many logs. By default it prints system logs.\n" +
		"Use --stdout and --stderr to get executor logs.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		t, err := client.GetTask(host, args[0])
		exitOnErr(err)

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
	logsCmd.Flags().BoolVar(&stdout, "stdout", false, "Prints executor standard out")
	logsCmd.Flags().BoolVar(&stderr, "stderr", false, "Prints executor standard error")
	rootCmd.AddCommand(logsCmd)
}

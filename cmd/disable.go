package cmd

import (
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cancel bool

var disableCmd = &cobra.Command{
	Use:     "disable id...",
	Aliases: []string{"remove", "rm"},
	Short:   "Disable or more worker nodes",
	Long: "Tasks will keep running at disabled node but no tasks will be submitted to worker.\n" +
		"Cancel option tells master server to cancel all tasks in node and enqueue those tasks.\n" +
		"It will print IDs of successfully disabled nodes.",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		for _, id := range args {
			if err := client.DisableNode(host, id, cancel); err != nil {
				message("Unable to disable worker %s: %s\n", id, err)
				continue
			}
			println(id)
		}
	},
}

func init() {
	disableCmd.Flags().BoolVarP(&cancel, "cancel", "c", false, "Cancel running tasks in worker node.")
	rootCmd.AddCommand(disableCmd)
}

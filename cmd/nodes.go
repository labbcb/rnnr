package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "ListTasks computing nodes",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		ns, err := client.ListNodes(host)
		fatalOnErr(err)

		var status string
		fmt.Printf("%-29s   %s   %-8s   %s\n", "Resources", "Tasks", "Status", "Host (port)")
		for _, n := range ns {
			if n.Active {
				status = "ACTIVE"
			} else {
				status = "INACTIVE"
			}
			fmt.Printf("CPU=%02d/%02d RAM=%06.2f/%06.2fGB | %02d    | %-8s | %s (%s)\n",
				n.Usage.CPUCores, n.CPUCores, n.Usage.RAMGb, n.RAMGb, n.Usage.Tasks, status, n.Host, n.Port)
		}
	},
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}

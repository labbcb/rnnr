package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List computing nodes",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		ns, err := client.ListNodes(host)
		fatalOnErr(err)

		var status string
		fmt.Printf("%-36s   %-29s   %s   %-8s   %s\n", "ID", "Resources", "Tasks", "Status", "URL")
		for _, n := range ns {
			if n.Active {
				status = "ACTIVE"
			} else {
				status = "INACTIVE"
			}
			fmt.Printf("%-36s | CPU=%02d/%02d RAM=%06.2f/%06.2fGB | %02d    | %-8s | %s\n",
				n.ID, n.Usage.CPUCores, n.Info.CPUCores, n.Usage.RAMGb, n.Info.RAMGb, n.Usage.Tasks, status, n.Host)
		}
	},
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}

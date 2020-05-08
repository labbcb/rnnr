package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var onlyActive bool

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List worker nodes",
	Long: "It will print worker nodes with number of active tasks and resource information.\n" +
		"Use --active to print only enabled nodes.\n" +
		"Use --format json to print in JSON format.",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		nodes, err := client.ListNodes(host, onlyActive)
		exitOnErr(err)

		format := viper.GetString("format")
		if format == "json" {
			exitOnErr(json.NewEncoder(os.Stdout).Encode(nodes))
			return
		}

		var status string
		fmt.Printf("%-29s   %s   %-8s   %s\n", "Resources", "Tasks", "Status", "Host (port)")
		for _, n := range nodes {
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
	nodesCmd.Flags().BoolVarP(&onlyActive, "active", "a", false, "Get only active nodes.")
	rootCmd.AddCommand(nodesCmd)
}

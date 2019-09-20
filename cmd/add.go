package cmd

import (
	"fmt"
	"log"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/node"
	"github.com/spf13/cobra"
)

func init() {
	addCmd.Flags().StringVar(&host, "host", "http://localhost:8080", "URL to RNNR server")
	addCmd.Flags().IntVar(&cpuCores, "cpu", 0, "Maximum CPU cores")
	addCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add NODE_URL",
	Short: "Add a computing note",
	Long: `It adds a computing node to RNNR master server which will check if the worker is running.
	This command can be used to change the maximum resources without restarting the worker node.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n := &node.Node{
			Host: args[0],
			Info: &node.Info{
				CPUCores: cpuCores,
				RAMGb:    ramGb,
			},
		}
		id, err := client.Add(host, n)
		if err != nil {
			log.Fatalf("Unable to activate node %s on %s: %v\n", n.Host, host, err)
		}
		fmt.Println(id)
	},
}

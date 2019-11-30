package cmd

import (
	"fmt"
	"log"

	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/node"
	"github.com/spf13/cobra"
)

var cpuCores int
var ramGb float64

var enableCmd = &cobra.Command{
	Use:     "enable url...",
	Aliases: []string{"add", "set"},
	Short:   "Enable one or more computing notes by their URL",
	Long: `It adds one or more computing nodes to RNNR master server which will check if the worker is running.
	By default it will not set any maximum CPU or Memory resources which implies to use all available in the node.
	This command can be used to change the maximum resources without restarting the worker node.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, url := range args {
			n := &node.Node{
				Host: url,
				Info: &node.Info{
					CPUCores: cpuCores,
					RAMGb:    ramGb,
				},
			}
			id, err := client.EnableNode(host, n)
			if err != nil {
				log.Printf("Unable to activate node %s on %s: %v\n", n.Host, host, err)
			}
			fmt.Println(id)
		}
	},
}

func init() {
	enableCmd.Flags().IntVar(&cpuCores, "cpu", 0, "Maximum CPU cores")
	enableCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	rootCmd.AddCommand(enableCmd)
}
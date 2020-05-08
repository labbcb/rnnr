package cmd

import (
	"fmt"
	"github.com/labbcb/rnnr/client"
	"github.com/labbcb/rnnr/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cpuCores int32
var ramGb float64

var enableCmd = &cobra.Command{
	Use:     "enable hostname",
	Aliases: []string{"add", "set"},
	Short:   "Enable a worker node",
	Long: "To add a RNNR worker node provide the hostname. It will be used as node ID.\n" +
		"It will guess the maximum CPU cores and memory in gigabytes.\n" +
		"Use --cpu and --ram to change these values.\n" +
		"It is recommended to not use all available computing resources.\n" +
		"Default port is 50051. Use --port to change this value.\n" +
		"Running this command for a already enabled node will update maximum resources.\n",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.EnableNode(host, &models.Node{
			Host:     args[0],
			Port:     port,
			CPUCores: cpuCores,
			RAMGb:    ramGb,
		})
		exitOnErr(err)
		fmt.Println(resp)
	},
}

func init() {
	enableCmd.Flags().StringVarP(&port, "port", "p", "50051", "Port of worker instance")
	enableCmd.Flags().Int32Var(&cpuCores, "cpu", 0, "Maximum CPU cores")
	enableCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	rootCmd.AddCommand(enableCmd)
}

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
	Use:     "enable worker",
	Aliases: []string{"add", "set"},
	Short:   "Enable a worker server",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		resp, err := client.EnableNode(host, &models.Node{
			Host:     args[0],
			Port:     port,
			CPUCores: cpuCores,
			RAMGb:    ramGb,
		})
		fatalOnErr(err)
		fmt.Println(resp)
	},
}

func init() {
	enableCmd.Flags().StringVarP(&port, "port", "p", "50051", "TCP port of worker server")
	enableCmd.Flags().Int32Var(&cpuCores, "cpu", 0, "Maximum CPU cores")
	enableCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	rootCmd.AddCommand(enableCmd)
}

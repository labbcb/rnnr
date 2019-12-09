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
		for _, n := range ns {
			fmt.Println(n)
		}
	},
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}

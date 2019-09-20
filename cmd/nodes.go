package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/labbcb/rnnr/client"
)

func init() {
	nodesCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8080", "URL to RNNR server")
	rootCmd.AddCommand(nodesCmd)
}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List computing nodes",
	Run: func(cmd *cobra.Command, args []string) {
		ns, err := client.ListNodes(host)
		if err != nil {
			fmt.Println("Unable to list nodes:", err)
		}

		for _, n := range ns {
			fmt.Println(n)
		}
	},
}

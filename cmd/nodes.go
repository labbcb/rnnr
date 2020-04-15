package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/labbcb/rnnr/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List computing nodes",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		nodes, err := client.ListNodes(host)
		fatalOnErr(err)

		format := viper.GetString("format")
		if format == "json" {
			if err := json.NewEncoder(os.Stdout).Encode(nodes); err != nil {
				log.Fatal(err)
			}
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
	rootCmd.AddCommand(nodesCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var port int
var volumes []string
var version string

var deployCmd = &cobra.Command{
	Use:   "deploy master|worker ...",
	Short: "Deploy RNNR as Docker container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]
		if mode != "master" && mode != "worker" {
			log.Fatalf("Invalid mode %s. Either master or worker.", mode)
		}

		if mode == "worker" && volumes == nil {
			log.Fatal("Provide at least one volume to be mounted in RNNR Worker container.")
		}

		//d, err := docker.Connect()
		//fatalOnErr(err)

		//id, err := d.Deploy(port, volumes, mode, version)
		//fatalOnErr(err)
		//println(id)
	},
}

func init() {
	deployCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to publish RNNR.")
	deployCmd.Flags().StringArrayVarP(&volumes, "volume", "v", nil, "Directories to mount inside container.")
	deployCmd.Flags().StringVar(&version, "version", "latest", "Docker tag of RNNR image.")
	rootCmd.AddCommand(deployCmd)
}

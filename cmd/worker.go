package cmd

import (
	"net"

	"github.com/labbcb/rnnr/pb"
	"github.com/labbcb/rnnr/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var port string

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"start"},
	Short:   "Start worker server",
	Long: "Start RNNR worker server instance.\n" +
		"It will listen port 50051 by default.\n" +
		"Use --port to change this value.\n" +
		"It requires access to Docker socket.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		w, err := worker.New(cpuCores, ramGb)
		exitOnErr(err)

		lis, err := net.Listen("tcp", ":"+port)
		exitOnErr(err)

		server := grpc.NewServer()
		pb.RegisterWorkerServer(server, w)
		exitOnErr(server.Serve(lis))
	},
}

func init() {
	workerCmd.Flags().StringVarP(&port, "port", "p", "50051", "Port to bind server")
	rootCmd.AddCommand(workerCmd)
}

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
	Short:   "Start RNNR worker server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		w, err := worker.New(cpuCores, ramGb)
		fatalOnErr(err)

		lis, err := net.Listen("tcp", ":"+port)
		fatalOnErr(err)

		server := grpc.NewServer()
		pb.RegisterWorkerServer(server, w)
		fatalOnErr(server.Serve(lis))
	},
}

func init() {
	workerCmd.Flags().StringVarP(&port, "port", "p", "50051", "TCP port to bind server")
	rootCmd.AddCommand(workerCmd)
}

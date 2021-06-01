package cmd

import (
	"net"

	"github.com/labbcb/rnnr/proto"
	"github.com/labbcb/rnnr/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var port, user, group string
var cpuCores int32
var ramGb float64
var volumes []string

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

		w, err := server.NewWorker(cpuCores, ramGb, volumes, user, group)
		exitOnErr(err)

		if w.Info.CpuCores > w.Info.IdentifiedCpuCores {
			log.Warnf("Defined number of CPU cores (%d) is greater than identified (%d).", w.Info.CpuCores, w.Info.IdentifiedCpuCores)
		}

		if w.Info.RamGb > w.Info.IdentifiedRamGb {
			log.Warnf("Defined number of RAM (%.2f GB) is greater than identified (%.2f GB).", w.Info.RamGb, w.Info.IdentifiedRamGb)
		}

		lis, err := net.Listen("tcp", ":"+port)
		exitOnErr(err)

		server := grpc.NewServer()
		proto.RegisterWorkerServer(server, w)
		exitOnErr(server.Serve(lis))
	},
}

func init() {
	workerCmd.Flags().StringVarP(&port, "port", "p", "50051", "Port to bind server")
	workerCmd.Flags().Int32Var(&cpuCores, "cpu", 0, "Maximum CPU cores")
	workerCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	workerCmd.Flags().StringArrayVarP(&volumes, "volume", "v", []string{}, "Volumes to mount in containers")
	workerCmd.Flags().StringVarP(&user, "user", "u", "root", "User name or UID")
	workerCmd.Flags().StringVarP(&group, "group", "g", "root", "Group name or GID")
	rootCmd.AddCommand(workerCmd)
}

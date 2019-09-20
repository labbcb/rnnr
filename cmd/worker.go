package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labbcb/rnnr/worker"
	"github.com/spf13/cobra"
)

func init() {
	workerCmd.Flags().StringVar(&database, "database", "mongodb://localhost:27017", "URL to Mongo database")
	workerCmd.Flags().StringVar(&address, "address", ":8080", "Addres to bind server")
	workerCmd.Flags().IntVar(&cpuCores, "cpu", 0, "Maximum CPU cores")
	workerCmd.Flags().Float64Var(&ramGb, "ram", 0, "Maximum memory in gigabytes")
	rootCmd.AddCommand(workerCmd)
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start RNNR standalone worker",
	Long: `Worker mode is a standalone version of RNNR server that uses Docker to execute tasks.
	It can be connected to RNNR master server to accept remote tasks (see rnnr add).
	If maximum CPU cores or memory is not define it it will guess the values.
	These values can be chenged through rnnr add.`,
	Run: func(cmd *cobra.Command, args []string) {
		w, err := worker.New(database, cpuCores, ramGb)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to create worker:", err)
			os.Exit(1)
		}

		http.ListenAndServe(address, w.Server.Router)
	},
}

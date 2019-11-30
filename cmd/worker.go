package cmd

import (
	"github.com/labbcb/rnnr/worker"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"start"},
	Short:   "Start RNNR standalone worker",
	Long: `Worker mode is a standalone version of RNNR server that uses Docker to execute tasks.
	It can be connected to RNNR master server to accept remote tasks (see rnnr add).
	Maximum CPU cores and memory are guessed.
	These values can be changed through rnnr enable.`,
	Run: func(cmd *cobra.Command, args []string) {
		w, err := worker.New(database, cpuCores, ramGb)
		fatalOnErr(err)

		log.Fatal(http.ListenAndServe(address, w.Server.Router))
	},
}

func init() {
	workerCmd.Flags().StringVar(&database, "database", "mongodb://localhost:27017", "URL to Mongo database")
	workerCmd.Flags().StringVar(&address, "address", ":8080", "Address to bind server")
	rootCmd.AddCommand(workerCmd)
}

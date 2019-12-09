package cmd

import (
	"github.com/labbcb/rnnr/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"start"},
	Short:   "StartMonitor RNNR standalone worker",
	Long: `Worker mode is a standalone version of RNNR server that uses Docker to execute tasks.
	It can be connected to RNNR master server to accept remote tasks (see rnnr add).
	Maximum CPU cores and memory are guessed.
	These values can be changed through rnnr enable.`,
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database")
		w, err := worker.New(database, cpuCores, ramGb)
		fatalOnErr(err)

		address := viper.GetString("address")
		log.Fatal(http.ListenAndServe(address, w.Server.Router))
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

package cmd

import (
	"net/http"
	"time"

	"github.com/labbcb/rnnr/master"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var database, address string
var sleepTime int

var masterCmd = &cobra.Command{
	Use:     "master",
	Aliases: []string{"server"},
	Short:   "StartMonitor RNNR master server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		m, err := master.New(database, time.Duration(sleepTime)*time.Second)
		fatalOnErr(err)

		log.Fatal(http.ListenAndServe(address, m.Router))
	},
}

func init() {
	masterCmd.PersistentFlags().StringVarP(&database, "database", "d", "mongodb://localhost:27017", "URL to Mongo database")
	masterCmd.PersistentFlags().StringVarP(&address, "address", "a", ":8080", "Address to bind server")
	masterCmd.Flags().IntVarP(&sleepTime, "time", "t", 5, "Sleep time in second for monitoring system.")
	rootCmd.AddCommand(masterCmd)
}

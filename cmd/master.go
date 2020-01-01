package cmd

import (
	"github.com/labbcb/rnnr/master"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

var database, address string

var masterCmd = &cobra.Command{
	Use:     "master",
	Aliases: []string{"server"},
	Short:   "StartMonitor RNNR master server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		m, err := master.New(database)
		fatalOnErr(err)

		log.Fatal(http.ListenAndServe(address, m.Router))
	},
}

func init() {
	masterCmd.PersistentFlags().StringVarP(&database, "database", "d", "mongodb://localhost:27017", "URL to Mongo database")
	masterCmd.PersistentFlags().StringVarP(&address, "address", "a", ":8080", "Address to bind server")
	rootCmd.AddCommand(masterCmd)
}

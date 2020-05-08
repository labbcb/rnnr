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
	Short:   "Start master server",
	Long: "Start the RNNR master server instance.\n" +
		"It will listen port 8080. Use --address to change the port suffixed with colon.\n" +
		"It will connect with MongoDB. use --database to change URL.\n" +
		"By default monitoring system will iterate over tasks and sleep. Use --time to change sleep time.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		m, err := master.New(database, time.Duration(sleepTime)*time.Second)
		exitOnErr(err)

		log.Fatal(http.ListenAndServe(address, m.Router))
	},
}

func init() {
	masterCmd.PersistentFlags().StringVarP(&database, "database", "d", "mongodb://localhost:27017", "URL to Mongo database")
	masterCmd.PersistentFlags().StringVarP(&address, "address", "a", ":8080", "Address to bind server")
	masterCmd.Flags().IntVarP(&sleepTime, "time", "t", 5, "Sleep time in second for monitoring system.")
	rootCmd.AddCommand(masterCmd)
}

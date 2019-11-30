package cmd

import (
	"github.com/labbcb/rnnr/master"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var masterCmd = &cobra.Command{
	Use:     "master",
	Aliases: []string{"server"},
	Short:   "Start RNNR master server",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := master.New(database)
		fatalOnErr(err)
		log.Fatal(http.ListenAndServe(address, m.Server.Router))
	},
}

func init() {
	masterCmd.Flags().StringVar(&database, "database", "mongodb://localhost:27017", "URL to Mongo database")
	masterCmd.Flags().StringVar(&address, "address", ":8080", "Address to bind server")
	rootCmd.AddCommand(masterCmd)
}

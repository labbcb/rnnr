package cmd

import (
	"github.com/labbcb/rnnr/master"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

var masterCmd = &cobra.Command{
	Use:     "master",
	Aliases: []string{"server"},
	Short:   "StartMonitor RNNR master server",
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database")
		m, err := master.New(database)
		fatalOnErr(err)

		address := viper.GetString("address")
		log.Fatal(http.ListenAndServe(address, m.Server.Router))
	},
}

func init() {
	rootCmd.AddCommand(masterCmd)
}

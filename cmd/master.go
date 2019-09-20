package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/labbcb/rnnr/master"
)

func init() {
	masterCmd.Flags().StringVar(&masterDb, "database", "mongodb://localhost:27017", "URL to Mongo database")
	masterCmd.Flags().StringVar(&masterAddr, "address", ":8080", "Addres to bind server")
	rootCmd.AddCommand(masterCmd)
}

var masterDb string
var masterAddr string
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Start RNNR master server",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := master.New(masterDb)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to start master server:", err)
			os.Exit(1)
		}
		http.ListenAndServe(masterAddr, m.Server.Router)
	},
}

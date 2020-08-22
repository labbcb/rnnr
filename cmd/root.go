// Package cmd implements command line commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "rnnr",
	Short: "Distributed task executor for genomics research",
	Long: "RNNR is designed to distribute processing tasks across computing nodes.\n" +
		"It implements Task Execution Service API to communicate with workflow managers.\n" +
		"Full documentation at https://bcblab.org/rnnr",
}

// Execute calls root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.rnnr.yaml)")
	rootCmd.PersistentFlags().String("host", "http://localhost:8080", "RNNR server URL")
	rootCmd.PersistentFlags().StringP("format", "f", "console", "Output format. JSON or console")
	exitOnErr(viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host")))
	exitOnErr(viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format")))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".wf" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".rnnr")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func message(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func messageAndExit(format string, a ...interface{}) {
	message(format, a...)
	os.Exit(1)
}

func exitOnErr(err error) {
	if err != nil {
		messageAndExit("%v\n", err)
	}
}

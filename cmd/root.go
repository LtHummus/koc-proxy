package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/lthummus/koc-proxy/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&disableDatabase, "disable-databases", false, "disable databases [for testing only]")
	rootCmd.PersistentFlags().BoolVarP(&verboseLogging, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().MarkHidden("disable-databases")
	rootCmd.AddCommand(proxyCommand, discordCmd, migrateCommand, redisTest, kickCommand, webCmd, portTestCommand, passwdCmd, listCmd)
}

var cfgFile string
var disableDatabase bool
var verboseLogging bool

var rootCmd = &cobra.Command{
	Use:   "kocproxy",
	Short: "A tool for running an authentication layer for KnockoutCity private servers.",
}

func initConfig() {
	if verboseLogging {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Warn().Msg("verbose logging enabled")
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		util.FatalIfError(err, "could not get home directory")

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kocproxy")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Str("config_file_path", viper.ConfigFileUsed()).Msg("could not read config file")
	}

	log.Info().Str("config_file_path", viper.ConfigFileUsed()).Msg("loaded config")

	if disableDatabase {
		log.Warn().Msg("disable-databases set")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

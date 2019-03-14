package cmd

import (
	"bytes"
	"io/ioutil"
	//"time"

	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string
var version string

var rootCmd = &cobra.Command{
	Use:   "lora-tdma-server",
	Short: "LoRa Server project tdma-server",
	Long: `LoRa TDMA Server is an open-source tdma-server, part of the LoRa Server project
	> source & copyright information: https://github.com/lioneie/lora-app-server`,
	//> documentation & support: https://www.loraserver.io/lora-app-server
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 5, "debug=5, info=4, error=2, fatal=1, panic=0")

	// bind flag to config vars
	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// defaults
	viper.SetDefault("general.password_hash_iterations", 100000)
	viper.SetDefault("tdma_server.bind", "0.0.0.0:5555")
	viper.SetDefault("postgresql.dsn", "postgres://loraserver_ts:dbpassword@127.0.0.1/loraserver_ts?sslmode=disable")
	viper.SetDefault("postgresql.automigrate", true)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

// Execute executes the root command.
func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		viper.SetConfigType("toml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("lora-tdma-server")
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults.")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}

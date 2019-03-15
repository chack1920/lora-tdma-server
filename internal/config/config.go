package config

import (
	//"time"

	//"github.com/gomodule/redigo/redis"
	"github.com/lioneie/lora-tdma-server/internal/common"
	//"github.com/lioneie/lora-app-server/internal/handler"
	//"github.com/lioneie/lora-app-server/internal/handler/gcppubsub"
	//"github.com/lioneie/lora-app-server/internal/handler/mqtthandler"
	"github.com/lioneie/lora-tdma-server/internal/asclient"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel               int `mapstructure:"log_level"`
		PasswordHashIterations int `mapstructure:"password_hash_iterations"`
	}

	PostgreSQL struct {
		DSN         string `mapstructure:"dsn"`
		Automigrate bool
		DB          *common.DBLogger `mapstructure:"db"`
	} `mapstructure:"postgresql"`

	TdmaServer struct {
		Bind string
	} `mapstructure:"tdma_server"`

	AppServer struct {
		Pool asclient.Pool
	} `mapstructure:"app_server"`
}

// C holds the global configuration.
var C Config

package config

import (
	"time"

	"github.com/gomodule/redigo/redis"
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

	Redis struct {
		URL         string        `mapstructure:"url"`
		MaxIdle     int           `mapstructure:"max_idle"`
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`
		Pool        *redis.Pool
	}

	TdmaServer struct {
		Bind      string
		Scheduler struct {
			SchedulerInterval time.Duration `mapstructure:"scheduler_interval"`
		} `mapstructure:"scheduler"`
	} `mapstructure:"tdma_server"`

	AppServer struct {
		Bind string `mapstructure:"bind"`
		Pool asclient.Pool
	} `mapstructure:"app_server"`
}

// C holds the global configuration.
var C Config

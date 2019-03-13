package config

import (
//"time"

//"github.com/gomodule/redigo/redis"
//"github.com/lioneie/lora-app-server/internal/common"
//"github.com/lioneie/lora-app-server/internal/handler"
//"github.com/lioneie/lora-app-server/internal/handler/gcppubsub"
//"github.com/lioneie/lora-app-server/internal/handler/mqtthandler"
//"github.com/lioneie/lora-app-server/internal/nsclient"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel               int `mapstructure:"log_level"`
		PasswordHashIterations int `mapstructure:"password_hash_iterations"`
	}
	TdmaServer struct {
		Bind string
	} `mapstructure:"tdma_server"`
}

// C holds the global configuration.
var C Config

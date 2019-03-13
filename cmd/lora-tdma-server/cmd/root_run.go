package cmd

import (
	"context"
	//"crypto/tls"
	//"crypto/x509"
	//"fmt"
	//"io/ioutil"
	//"net"
	//"net/http"
	"os"
	"os/signal"
	//"strings"
	"syscall"
	//"time"

	//assetfs "github.com/elazarl/go-bindata-assetfs"
	//"github.com/gofrs/uuid"
	//"github.com/gorilla/mux"
	//"github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	//"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	//"github.com/grpc-ecosystem/grpc-gateway/runtime"
	//"github.com/pkg/errors"
	//migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	//"github.com/tmc/grpc-websocket-proxy/wsproxy"
	//"golang.org/x/net/http2"
	//"golang.org/x/net/http2/h2c"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/credentials"
	//pb "github.com/lioneie/lora-app-server/api"
	//"github.com/lioneie/lora-app-server/internal/api"
	//"github.com/lioneie/lora-app-server/internal/api/auth"
	"github.com/lioneie/lora-tdma-server/internal/config"
	//"github.com/lioneie/lora-app-server/internal/downlink"
	//"github.com/lioneie/lora-app-server/internal/gwping"
	//"github.com/lioneie/lora-app-server/internal/handler"
	//"github.com/lioneie/lora-app-server/internal/handler/gcppubsub"
	//"github.com/lioneie/lora-app-server/internal/handler/mqtthandler"
	//"github.com/lioneie/lora-app-server/internal/handler/multihandler"
	//"github.com/lioneie/lora-app-server/internal/migrations"
	//"github.com/lioneie/lora-app-server/internal/nsclient"
	//"github.com/lioneie/lora-app-server/internal/static"
	//"github.com/lioneie/lora-app-server/internal/storage"
	//"github.com/lioneie/loraserver/api/as"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		printStartMessage,
		startTdmaServerAPI,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping lora-app-server")
		// todo: handle graceful shutdown?
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}

func setLogLevel() error {
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	return nil
}

func printStartMessage() error {
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://www.loraserver.io/",
	}).Info("starting LoRa TDMA Server")
	return nil
}

func startTdmaServerAPI() error {
	return nil
}

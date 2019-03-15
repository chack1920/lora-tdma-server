package cmd

import (
	"context"
	//"crypto/tls"
	//"crypto/x509"
	"fmt"
	//"io/ioutil"
	//"net"
	"net/http"
	"os"
	"os/signal"
	//"strings"
	"syscall"
	"time"

	//assetfs "github.com/elazarl/go-bindata-assetfs"
	//"github.com/gofrs/uuid"
	//"github.com/gorilla/mux"
	//"github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	//"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	//"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	//migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	//"github.com/tmc/grpc-websocket-proxy/wsproxy"
	//"golang.org/x/net/http2"
	//"golang.org/x/net/http2/h2c"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/credentials"
	//pb "github.com/lioneie/lora-app-server/api"
	"github.com/lioneie/lora-tdma-server/internal/api"
	//"github.com/lioneie/lora-app-server/internal/api/auth"
	"github.com/lioneie/lora-tdma-server/internal/config"
	//"github.com/lioneie/lora-app-server/internal/downlink"
	//"github.com/lioneie/lora-app-server/internal/gwping"
	//"github.com/lioneie/lora-app-server/internal/handler"
	//"github.com/lioneie/lora-app-server/internal/handler/gcppubsub"
	//"github.com/lioneie/lora-app-server/internal/handler/mqtthandler"
	//"github.com/lioneie/lora-app-server/internal/handler/multihandler"
	//"github.com/lioneie/lora-app-server/internal/migrations"
	"github.com/lioneie/lora-tdma-server/internal/asclient"
	//"github.com/lioneie/lora-app-server/internal/static"
	"github.com/lioneie/lora-tdma-server/internal/common"
	//"github.com/lioneie/loraserver/api/as"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/lioneie/lora-tdma-server/internal/mqttpubsub"
	"github.com/lioneie/lora-tdma-server/internal/multicast"
	"github.com/lioneie/lora-tdma-server/internal/storage"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		printStartMessage,
		startTdmaServerAPI,
		startMqttHandler,
		setRedisPool,
		setPostgreSQLConnection,
		setAppServerClient,
		testExample,
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
		log.Warning("stopping lora-tdma-server")
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
	log.WithFields(log.Fields{
		"bind": config.C.TdmaServer.Bind,
	}).Info("starting tdma-server api")

	server := http.Server{
		Handler: api.NewTdmaServerAPI(),
		Addr:    config.C.TdmaServer.Bind,
	}

	go func() {
		err := server.ListenAndServe()
		log.WithError(err).Error("tdma-server api error")
	}()
	return nil
}

func startMqttHandler() error {
	var pubsub *mqttpubsub.Backend
	for {
		var err error
		pubsub, err = mqttpubsub.NewBackend() //TODO:add config
		if err == nil {
			break
		}

		log.Errorf("could not setup mqtt backend, retry in 2 seconds: %s", err)
		time.Sleep(2 * time.Second)
	}
	//defer pubsub.Close()

	err := pubsub.SubscribeAppTopic()
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func setRedisPool() error {
	log.WithField("url", config.C.Redis.URL).Info("setup redis connection pool")
	config.C.Redis.Pool = common.NewRedisPool(
		config.C.Redis.URL,
		config.C.Redis.MaxIdle,
		config.C.Redis.IdleTimeout,
	)
	return nil
}

func setPostgreSQLConnection() error {
	log.Info("connecting to postgresql")
	db, err := common.OpenDatabase(config.C.PostgreSQL.DSN)
	if err != nil {
		return errors.Wrap(err, "database connection error")
	}
	config.C.PostgreSQL.DB = db
	return nil
}

func setAppServerClient() error {
	log.Info("set app server client")
	config.C.AppServer.Pool = asclient.NewPool()
	return nil
}

func testMulticastEnqueue() error {
	var mcData []byte = []byte{0xa, 0xb, 0xc, 0xd, 0xe}
	MulticastGroupId := "4a21c7f8-4111-4e46-97c9-2986ca60bac5"
	fcnt, err := multicast.Enqueue(MulticastGroupId, 5, mcData)
	if err == nil {
		log.Info("send multicaset success, fcnt: ", fcnt)
	} else {
		log.Error("send multicast err")
	}
	return nil
}

func testPostgreSQL() error {
	qi := storage.TestItem{
		FCnt: 1,
	}
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		return sqlxExt(tx, qi)
	})
	return err
}

func sqlxExt(db sqlx.Ext, qi storage.TestItem) error {
	if err := storage.CreateTestItem(db, &qi); err != nil {
		return errors.Wrap(err, "create test-item error")
	}
	return nil
}

func testRedis() error {
	var val []byte
	var err error

	p := config.C.Redis.Pool
	c := p.Get()
	key := "key1"
	exp := 30 * 1000

	val, err = redis.Bytes(c.Do("GET", key))
	fmt.Println("before set:", val, err)

	val = []byte{5, 6, 7}
	_, err = c.Do("PSETEX", key, exp, val)
	if err != nil {
		return errors.Wrap(err, "set redis error")
	}

	val, err = redis.Bytes(c.Do("GET", key))
	fmt.Println("after set:", val, err)
	return nil
}

func testExample() error {
	//return testMulticastEnqueue()
	//return testPostgreSQL
	return testRedis()
}

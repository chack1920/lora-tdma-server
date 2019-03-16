package cmd

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/lioneie/lora-tdma-server/internal/multicast"
	"github.com/lioneie/lora-tdma-server/internal/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func TestExample() error {
	return testMulticastEnqueue()
	//return testPostgreSQL
	//return testRedis()
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

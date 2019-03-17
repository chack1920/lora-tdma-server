package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	//"github.com/jmoiron/sqlx"
	"github.com/lioneie/lorawan"
	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
)

const (
	TdmaSessionKeyTempl = "lora:ts:tdma:session:%v"
)

type TdmaSessionItem struct {
	Time   time.Time
	DevEUI lorawan.EUI64
}

func GetTdmaSessionItemCache(p *redis.Pool, devEUI lorawan.EUI64) (TdmaSessionItem, error) {
	var dp TdmaSessionItem
	key := fmt.Sprintf(TdmaSessionKeyTempl, devEUI)

	c := p.Get()
	defer c.Close()

	val, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return dp, ErrDoesNotExist
		}
		return dp, errors.Wrap(err, "get error")
	}

	err = gob.NewDecoder(bytes.NewReader(val)).Decode(&dp)
	if err != nil {
		return dp, errors.Wrap(err, "gob decode error")
	}

	return dp, nil
}

func CreateTdmaSessionItemCache(p *redis.Pool, dp TdmaSessionItem) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(dp); err != nil {
		return errors.Wrap(err, "gob encode tdma session item error")
	}

	c := p.Get()
	defer c.Close()

	key := fmt.Sprintf(TdmaSessionKeyTempl, dp.DevEUI)

	_, err := c.Do("SET", key, buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "set tdma session error")
	}

	return nil
}

func FlushTdmaSessionItemCache(p *redis.Pool, devEUI lorawan.EUI64) error {
	key := fmt.Sprintf(TdmaSessionKeyTempl, devEUI)
	c := p.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return errors.Wrap(err, "delete error")
	}
	return nil
}

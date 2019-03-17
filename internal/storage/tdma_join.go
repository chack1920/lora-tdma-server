package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	//"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/lioneie/lorawan"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	TdmaJoinItemKeyTempl = "lora:ts:tj:%v"
)

type TdmaJoinItem struct {
	ID        int64         `db:"id"`
	CreatedAt time.Time     `db:"created_at"`
	DevEUI    lorawan.EUI64 `db:"dev_eui"`
	McSeq     uint8         `db:"mc_seq"`
	TxCycle   uint32        `db:"tx_cycle"`
}

func CreateTdmaJoinItem(db sqlx.Queryer, item *TdmaJoinItem) error {
	item.CreatedAt = time.Now()

	err := sqlx.Get(db, &item.ID, `
                insert into tdma_join (
                        created_at,
			dev_eui,
			mc_seq,
			tx_cycle
                ) values ($1, $2, $3, $4)
                returning
                        id
                `,
		item.CreatedAt,
		item.DevEUI,
		item.McSeq,
		item.TxCycle,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":       item.ID,
		"dev_eui":  item.DevEUI,
		"mc_seq":   item.McSeq,
		"tx_cycle": item.TxCycle,
	}).Info("join-tdma-item created")

	return nil
}

func DeleteTdmaJoinItem(db sqlx.Execer, devEUI lorawan.EUI64) error {
	res, err := db.Exec("delete from tdma_join where dev_eui = $1", devEUI[:])
	if err != nil {
		return handlePSQLError(err, "delete error")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return handlePSQLError(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithField("dev_eui", devEUI).Info("tdma join item deleted")
	return nil
}

func UpdateTdmaJoinItem(db sqlx.Queryer, item *TdmaJoinItem) (uint8, error) {
	var mcSeq uint8
	item.CreatedAt = time.Now()
	err := sqlx.Get(db, &mcSeq, `
                update
                        tdma_join
                set
                        created_at = $1,
			tx_cycle = $3
                where
                        dev_eui = $2
                returning
                        mc_seq
		`,
		item.CreatedAt,
		item.DevEUI,
		item.TxCycle,
	)
	if err != nil {
		return mcSeq, handlePSQLError(err, "update error")
	}

	log.WithFields(log.Fields{
		"dev_eui":  item.DevEUI,
		"tx_cycle": item.TxCycle,
	}).Info("tdma join updated")

	return mcSeq, nil
}

func GetTdmaJoinItem(db sqlx.Queryer, devEUI lorawan.EUI64) (TdmaJoinItem, error) {
	var ret TdmaJoinItem
	err := sqlx.Get(db, &ret, `
                select
                        *
                from
                        tdma_join
                where
                        dev_eui = $1`,
		devEUI,
	)
	if err != nil {
		return ret, handlePSQLError(err, "select error")
	}
	return ret, nil
}

//TODO: will remove
func GetTdmaJoinItemCounter(db sqlx.Queryer) (int64, error) {
	var ret int64
	err := sqlx.Get(db, &ret, `
                select
                        count(*)
                from
                        tdma_join`,
	)
	if err != nil {
		return ret, handlePSQLError(err, "select error")
	}
	return ret, nil
}

func GetTdmaJoinItemCache(p *redis.Pool, devEUI lorawan.EUI64) (TdmaJoinItem, error) {
	var dp TdmaJoinItem
	key := fmt.Sprintf(TdmaJoinItemKeyTempl, devEUI)

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

func CreateTdmaJoinItemCache(p *redis.Pool, dp TdmaJoinItem) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(dp); err != nil {
		return errors.Wrap(err, "gob encode tdma join item error")
	}

	c := p.Get()
	defer c.Close()

	key := fmt.Sprintf(TdmaJoinItemKeyTempl, dp.DevEUI)

	_, err := c.Do("SET", key, buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "set tdma join error")
	}

	return nil
}

func FlushTdmaJoinItemCache(p *redis.Pool, devEUI lorawan.EUI64) error {
	key := fmt.Sprintf(TdmaJoinItemKeyTempl, devEUI)
	c := p.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return errors.Wrap(err, "delete error")
	}
	return nil
}

func GetAndCacheTdmaJoinItem(db sqlx.Queryer, p *redis.Pool, devEUI lorawan.EUI64) (TdmaJoinItem, error) {
	dp, err := GetTdmaJoinItemCache(p, devEUI)
	if err == nil {
		return dp, nil
	}

	if err != ErrDoesNotExist {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
		}).WithError(err).Error("get tdma join item cache error")
		// we don't return as we can still fall-back onto db retrieval
	}

	dp, err = GetTdmaJoinItem(db, devEUI)
	if err != nil {
		return TdmaJoinItem{}, errors.Wrap(err, "get tdma join error")
	}

	err = CreateTdmaJoinItemCache(p, dp)
	if err != nil {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
		}).WithError(err).Error("create tdma join item cache error")
	}

	return dp, nil
}

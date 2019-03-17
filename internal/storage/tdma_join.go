package storage

import (
	"time"

	//"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	//"github.com/pkg/errors"
	"github.com/lioneie/lorawan"
	log "github.com/sirupsen/logrus"
)

type TdmaJoinItem struct {
	ID        int64         `db:"id"`
	CreatedAt time.Time     `db:"created_at"`
	DevEUI    lorawan.EUI64 `db:"dev_eui"`
	McSeq     uint8         `db:"mc_seq"`
	TxCycle   uint32        `db:"tx_cycle"`
}

// CreateMulticastQueueItem adds the given item to the queue.
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

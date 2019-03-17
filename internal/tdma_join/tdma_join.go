package tdma_join

import (
	"github.com/jmoiron/sqlx"
	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/lioneie/lora-tdma-server/internal/storage"
	"github.com/lioneie/lorawan"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	getFailedMulticastSeq uint8 = 255
)

func HandleTdmaJoinRequest(pl lorawan.TdmaReqPayload) lorawan.TdmaAnsPayload {
	var err error
	var ans lorawan.TdmaAnsPayload
	var mcSeq uint8

	mcSeq, err = updateTdmaJoinItem(pl)

	if err != nil {
		mcSeq, err = createTdmaJoinItem(pl)
		if err != nil {
			mcSeq = getFailedMulticastSeq
		}
	}

	ans = lorawan.TdmaAnsPayload{
		DevEUI: pl.DevEUI,
		McSeq:  mcSeq,
	}
	log.WithFields(log.Fields{
		"dev_eui": ans.DevEUI,
		"mc_seq":  ans.McSeq,
	}).Info("tdma join answer")
	_ = storage.FlushTdmaJoinItemCache(config.C.Redis.Pool, pl.DevEUI)
	return ans
}

func updateTdmaJoinItem(pl lorawan.TdmaReqPayload) (uint8, error) {
	var mcSeq uint8
	var err error
	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		tmp := storage.TdmaJoinItem{
			DevEUI:  pl.DevEUI,
			TxCycle: uint32(pl.TxCycle),
		}
		if mcSeq, err = storage.UpdateTdmaJoinItem(tx, &tmp); err != nil {
			return errors.Wrap(err, "get tdma-join-item error")
		}
		return nil

	})

	return mcSeq, err
}

func createTdmaJoinItem(pl lorawan.TdmaReqPayload) (uint8, error) {
	var cnt int64 = 0
	var err error
	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		if cnt, err = storage.GetTdmaJoinItemCounter(tx); err != nil {
			return errors.Wrap(err, "get tdma-join-item counter error")
		}
		return nil
	})

	err = storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		tmp := storage.TdmaJoinItem{
			DevEUI:  pl.DevEUI,
			McSeq:   uint8(cnt),
			TxCycle: uint32(pl.TxCycle),
		}
		if err = storage.CreateTdmaJoinItem(tx, &tmp); err != nil {
			return errors.Wrap(err, "create tdma-join-item error")
		}
		return nil
	})
	return uint8(cnt), err
}

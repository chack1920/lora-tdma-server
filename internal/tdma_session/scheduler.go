package tdma_session

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/lioneie/lora-tdma-server/internal/storage"
)

func TdmaSessionSchedulerLoop() {
	for {
		log.Debug("running tdma session scheduler batch")
		if err := ScheduleTdmaSession(); err != nil {
			log.WithError(err).Error("tdma session scheduler error")
		}
		time.Sleep(config.C.TdmaServer.Scheduler.SchedulerInterval)
	}
}

func ScheduleTdmaSession() error {
	tdmaSessionItems, err := storage.GetSchedulableTdmaSessionItems(config.C.Redis.Pool)
	if err != nil {
		return errors.Wrap(err, "get tdma session error")
	}

	for _, item := range tdmaSessionItems {
		//TODO
		log.Info(item)
		calcNextTime(item)
	}

	return nil
}

func calcNextTime(item storage.TdmaSessionItem) {
	var tdmaJoinItem storage.TdmaJoinItem
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		var err error
		tdmaJoinItem, err = storage.GetAndCacheTdmaJoinItem(tx, config.C.Redis.Pool, item.DevEUI)
		return err
	})
	if err != nil {
		return
	}
	log.Debug(tdmaJoinItem.TxCycle)
	item.Time = time.Now().Add(time.Duration(tdmaJoinItem.TxCycle) * time.Millisecond)
	_ = storage.CreateTdmaSessionItemCache(config.C.Redis.Pool, item)
}

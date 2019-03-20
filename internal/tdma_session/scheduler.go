package tdma_session

import (
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/lioneie/lora-tdma-server/internal/multicast"
	"github.com/lioneie/lora-tdma-server/internal/storage"
)

var errAbort = errors.New("")

type tdmaSessionContext struct {
	item storage.TdmaSessionItem
}

type schedulerIntervalContext struct {
	//TODO
	num uint8
}

var tdmaSessionTasks = []func(*tdmaSessionContext) error{
	txMulticastData,
	calcNextTime,
	updateSchedulerContext,
}

var scheIntCtx schedulerIntervalContext

func TdmaSessionSchedulerLoop() {
	for {
		log.Debug("running tdma session scheduler batch")
		if err := ScheduleTdmaSession(); err != nil {
			log.WithError(err).Error("tdma session scheduler error")
		}
		calcIntervalAndSleep()
	}
}

func ScheduleTdmaSession() error {
	tdmaSessionItems, err := storage.GetSchedulableTdmaSessionItems(config.C.Redis.Pool)
	if err != nil {
		return errors.Wrap(err, "get tdma session error")
	}

	for _, item := range tdmaSessionItems {
		log.Debug(item)
		_ = handleTdmaSessionItem(item)
	}

	return nil
}

func handleTdmaSessionItem(item storage.TdmaSessionItem) error {
	ctx := tdmaSessionContext{
		item: item,
	}

	for _, t := range tdmaSessionTasks {
		if err := t(&ctx); err != nil {
			if err == errAbort {
				return nil
			}
			return err
		}
	}

	return nil
}

func txMulticastData(ctx *tdmaSessionContext) error {
	var mcData []byte = []byte{0xa, 0xb, 0xc, 0xd, 0xe}
	MulticastGroupId := "4a21c7f8-4111-4e46-97c9-2986ca60bac5" //TODO
	fcnt, err := multicast.Enqueue(MulticastGroupId, 5, mcData)
	if err == nil {
		log.Info("send multicaset success, fcnt: ", fcnt)
	} else {
		log.Error("send multicast err:", err)
	}
	return nil
}

func calcNextTime(ctx *tdmaSessionContext) error {
	item := ctx.item
	var tdmaJoinItem storage.TdmaJoinItem
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		var err error
		tdmaJoinItem, err = storage.GetAndCacheTdmaJoinItem(tx, config.C.Redis.Pool, item.DevEUI)
		return err
	})
	if err != nil {
		return err
	}
	log.Debug(tdmaJoinItem.TxCycle)
	item.Time = time.Now().Add(time.Duration(tdmaJoinItem.TxCycle) * time.Millisecond)
	_ = storage.CreateTdmaSessionItemCache(config.C.Redis.Pool, item)
	return nil
}

func updateSchedulerContext(ctx *tdmaSessionContext) error {
	//TODO
	scheIntCtx.num++
	return nil
}

func calcIntervalAndSleep() {
	var sleepTime time.Duration
	//TODO: SF12 max payload: 2794ms
	var timeOnAir int64 = 2794
	var ceilVal int64 = int64(math.Ceil(float64(scheIntCtx.num) / 8))
	sleepTime = time.Duration(timeOnAir*ceilVal + timeOnAir)
	sleepTime *= time.Millisecond

	if config.C.TdmaServer.Scheduler.SchedulerInterval > sleepTime {
		sleepTime = config.C.TdmaServer.Scheduler.SchedulerInterval
	}

	scheIntCtx.num = 0
	time.Sleep(config.C.TdmaServer.Scheduler.SchedulerInterval)
}

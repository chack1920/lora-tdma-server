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
	sessionItem storage.TdmaSessionItem
	joinItem    storage.TdmaJoinItem
}

type schedulerContext struct {
	maxMulticastSequence uint8
	totalCounter         uint8
	seqArray             []uint8
}

var tdmaSessionTasks = []func(*tdmaSessionContext) error{
	getTdmaJoinItem,
	calcNextTime,
	updateSchedulerContext,
}

var scheCtx schedulerContext

func TdmaSessionSchedulerLoop() {
	for {
		resetSchedulerContext()
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

	txMulticastData()

	return nil
}

func handleTdmaSessionItem(item storage.TdmaSessionItem) error {
	ctx := tdmaSessionContext{
		sessionItem: item,
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

func getTdmaJoinItem(ctx *tdmaSessionContext) error {
	var tdmaJoinItem storage.TdmaJoinItem
	err := storage.Transaction(config.C.PostgreSQL.DB, func(tx sqlx.Ext) error {
		var err error
		tdmaJoinItem, err = storage.GetAndCacheTdmaJoinItem(tx, config.C.Redis.Pool, ctx.sessionItem.DevEUI)
		return err
	})
	if err != nil {
		return err
	}
	ctx.joinItem = tdmaJoinItem
	return nil
}

func calcNextTime(ctx *tdmaSessionContext) error {
	log.Debug("calcNextTime, ctx.joinItem.TxCycle:", ctx.joinItem.TxCycle)
	ctx.sessionItem.Time = time.Now().Add(time.Duration(ctx.joinItem.TxCycle) * time.Millisecond)
	_ = storage.CreateTdmaSessionItemCache(config.C.Redis.Pool, ctx.sessionItem)
	return nil
}

func updateSchedulerContext(ctx *tdmaSessionContext) error {
	if ctx.joinItem.McSeq > scheCtx.maxMulticastSequence {
		scheCtx.maxMulticastSequence = ctx.joinItem.McSeq
	}
	scheCtx.seqArray = append(scheCtx.seqArray, ctx.joinItem.McSeq)
	scheCtx.totalCounter++
	return nil
}

func resetSchedulerContext() {
	scheCtx.totalCounter = 0
	scheCtx.maxMulticastSequence = 0
	scheCtx.seqArray = scheCtx.seqArray[0:0]
}

func calcIntervalAndSleep() {
	var sleepTime time.Duration = config.C.TdmaServer.Scheduler.SchedulerInterval
	if scheCtx.totalCounter > 0 {
		//TODO: SF12 max payload: 2794ms
		var timeOnAir int64 = 2794
		var ceilVal int64 = int64(math.Ceil(float64(scheCtx.totalCounter) / 8))
		sleepTime = time.Duration(timeOnAir*ceilVal + timeOnAir)
		sleepTime *= time.Millisecond

		if config.C.TdmaServer.Scheduler.SchedulerInterval > sleepTime {
			sleepTime = config.C.TdmaServer.Scheduler.SchedulerInterval
		}
	}

	time.Sleep(sleepTime)
}

func txMulticastData() {
	log.Debug("txMulticastData, scheCtx.totalCounter:", scheCtx.totalCounter)
	if scheCtx.totalCounter <= 0 {
		return
	}
	var dataLen uint8 = (scheCtx.maxMulticastSequence / 8) + 1
	var mcData []byte = make([]byte, dataLen, dataLen)
	var byteNum, bitNum uint8
	for _, val := range scheCtx.seqArray {
		byteNum = val / 8
		bitNum = val % 8
		mcData[byteNum] |= (1 << bitNum)
	}
	MulticastGroupId := "4a21c7f8-4111-4e46-97c9-2986ca60bac5" //TODO
	log.Debug("txMulticastData, mcData:", mcData)
	fcnt, err := multicast.Enqueue(MulticastGroupId, 5, mcData)
	if err == nil {
		log.Info("send multicaset success, fcnt: ", fcnt)
	} else {
		log.Error("send multicast err:", err)
	}
}

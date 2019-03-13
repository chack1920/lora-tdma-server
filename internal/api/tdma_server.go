package api

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/lioneie/lora-tdma-server/internal/tdma_join"
	"github.com/lioneie/lorawan"
	"github.com/lioneie/lorawan/backend"
)

type TdmaServerAPI struct{}

func NewTdmaServerAPI() http.Handler {
	return &TdmaServerAPI{}
}

// ServeHTTP implements the http.Handler interface.
func (a *TdmaServerAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var pl lorawan.TdmaReqPayload

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.returnError(w, http.StatusInternalServerError, backend.Other, "read body error")
		return
	}

	err = json.Unmarshal(b, &pl)
	if err != nil {
		a.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	log.WithFields(log.Fields{
		"DevEUI":  pl.DevEUI,
		"DevAddr": pl.DevAddr,
		"TxCycle": pl.TxCycle,
	}).Info("ts: request received")

	//TODO: for expanding
	a.handleTdmaJoinReq(w, b)
}

func (a *TdmaServerAPI) returnError(w http.ResponseWriter, code int, resultCode backend.ResultCode, msg string) {
	log.WithFields(log.Fields{
		"error": msg,
	}).Error("js: error handling request")

	w.WriteHeader(code)

	pl := backend.Result{
		ResultCode:  resultCode,
		Description: msg,
	}
	b, err := json.Marshal(pl)
	if err != nil {
		log.WithError(err).Error("marshal json error")
		return
	}

	w.Write(b)
}

func (a *TdmaServerAPI) handleTdmaJoinReq(w http.ResponseWriter, b []byte) {
	var pl lorawan.TdmaReqPayload
	err := json.Unmarshal(b, &pl)
	if err != nil {
		a.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	ans := tdma_join.HandleTdmaJoinRequest(pl)

	a.returnPayload(w, http.StatusOK, ans)
}

func (a *TdmaServerAPI) returnPayload(w http.ResponseWriter, code int, pl interface{}) {
	w.WriteHeader(code)

	b, err := json.Marshal(pl)
	if err != nil {
		log.WithError(err).Error("marshal json error")
		return
	}
	//TODO: add log

	w.Write(b)
}

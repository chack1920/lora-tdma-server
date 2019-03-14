package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lioneie/lorawan"
)

type TdmaServerAPI struct{}

// NewTdmaServerAPI create a new TdmaServerAPI.
func NewTdmaServerAPI() http.Handler {
	return &TdmaServerAPI{}
}

// ServeHTTP implements the http.Handler interface.
func (a *TdmaServerAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req lorawan.TdmaReqPayload

	reqb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("ioutil.ReadAll error")
		return
	}

	err = json.Unmarshal(reqb, &req)
	if err != nil {
		fmt.Println("json.Unmarshal error")
		return
	}

	fmt.Println("req:", req)

	var ans lorawan.TdmaAnsPayload = lorawan.TdmaAnsPayload{
		DevEUI: req.DevEUI,
		McSeq:  55, //TODO
	}

	ansb, err := json.Marshal(ans)
	if err != nil {
		fmt.Println("marshal json error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ansb)
}

func http_client() {
	var ans lorawan.TdmaAnsPayload
	var req lorawan.TdmaReqPayload = lorawan.TdmaReqPayload{
		DevEUI:  lorawan.EUI64{0, 0, 0, 0, 0, 0, 0, 5},
		DevAddr: lorawan.DevAddr{0, 0, 0, 5},
		TxCycle: 55,
	}

	reqb, err := json.Marshal(req)
	if err != nil {
		fmt.Println("json.Marshal error")
		return
	}

	svr_addr := "http://localhost:5555"
	ansb, err := http.Post(svr_addr, "application/json", bytes.NewReader(reqb))
	if err != nil {
		fmt.Println("http.Post error", err)
		return
	}
	defer ansb.Body.Close()

	err = json.NewDecoder(ansb.Body).Decode(&ans)
	if err != nil {
		fmt.Println("json.NewDecoder error")
		return
	}
	fmt.Println("ans:", ans)
}

func http_server() {
	addr := "0.0.0.0:5555"
	server := http.Server{
		Handler: NewTdmaServerAPI(),
		Addr:    addr,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("tdma-server api error")
		return
	}
}

func main() {
	//http_server()
	http_client()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	fmt.Println(<-sigChan, "signal received")
}

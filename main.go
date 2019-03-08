package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	pb "github.com/lioneie/lora-app-server/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/lioneie/lorawan"
)

const (
	address = "localhost:8080"
)

func enqueue_devq() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDeviceQueueServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	md := metadata.New(map[string]string{"authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJsb3JhLWFwcC1zZXJ2ZXIiLCJhdWQiOiJsb3JhLWFwcC1zZXJ2ZXIiLCJuYmYiOjE0ODk1NjY5NTgsImV4cCI6MTU3MzQ0MTg3MSwic3ViIjoidXNlciIsInVzZXJuYW1lIjoiYWRtaW4ifQ.pFiL1wew4_fXMUmhkNNpuyX0VqiV0L3dcpZKgcsXoEA"})
	ctx = metadata.NewIncomingContext(ctx, md)
	//md, ok := metadata.FromIncomingContext(ctx)
	//fmt.Print(md, ok)

	//var base64_str string = "abc555"
	//var dl_data []byte = []byte(base64_str)
	var dl_data []byte = []byte{5, 6, 7, 8}
	r, err := c.Enqueue(ctx, &pb.EnqueueDeviceQueueItemRequest{
		DeviceQueueItem: &pb.DeviceQueueItem{
			DevEui:    "0000000000000020",
			Confirmed: false,
			FPort:     5,
			Data:      dl_data,
		},
	})
	if err != nil {
		log.Fatalf("send downlink error: %v", err)
	}
	log.Printf("enqueue downlink success, fcnt: %s", r.FCnt)
}

func enqueue_multicastq() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMulticastGroupServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	md := metadata.New(map[string]string{"authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJsb3JhLWFwcC1zZXJ2ZXIiLCJhdWQiOiJsb3JhLWFwcC1zZXJ2ZXIiLCJuYmYiOjE0ODk1NjY5NTgsImV4cCI6MTU3MzQ0MTg3MSwic3ViIjoidXNlciIsInVzZXJuYW1lIjoiYWRtaW4ifQ.pFiL1wew4_fXMUmhkNNpuyX0VqiV0L3dcpZKgcsXoEA"})
	ctx = metadata.NewIncomingContext(ctx, md)
	//md, ok := metadata.FromIncomingContext(ctx)
	//fmt.Print(md, ok)

	//var base64_str string = "abc555"
	//var dl_data []byte = []byte(base64_str)
	var mc_data []byte = []byte{0xa, 0xb, 0xc, 0xd, 0xe}
	r, err := c.Enqueue(ctx, &pb.EnqueueMulticastQueueItemRequest{
		MulticastQueueItem: &pb.MulticastQueueItem{
			MulticastGroupId: "4a21c7f8-4111-4e46-97c9-2986ca60bac5",
			FPort:            5,
			Data:             mc_data,
		},
	})
	if err != nil {
		log.Fatalf("send multicast error: %v", err)
	}
	log.Printf("enqueue multicast success, fcnt: %s", r.FCnt)
}

var mqtt_msg_handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func sub_mqtt() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("lr_paper_tdma")
	opts.SetKeepAlive(2 * time.Second)
	//opts.SetDefaultPublishHandler(mqtt_msg_handler)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	var topic = "application/1/#"
	var qos byte = 0
	var callback = mqtt_msg_handler
	if token := c.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	fmt.Println(<-sigChan, "signal received")
}

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

func http_server(wg *sync.WaitGroup) {
	defer wg.Done()
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
	var wg sync.WaitGroup
	wg.Add(1) //TODO
	//enqueue_devq()
	//enqueue_multicastq()
	//sub_mqtt()

	go http_server(&wg)
	//http_client()
	wg.Wait()
}

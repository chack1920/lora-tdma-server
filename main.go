package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/brocaar/lr_paper_tdma/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/eclipse/paho.mqtt.golang"
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

func main() {
	//enqueue_devq()
	//enqueue_multicastq()
	sub_mqtt()
}

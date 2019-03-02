package main

import (
	"context"
	//	"fmt"
	"log"
	"time"

	pb "github.com/brocaar/lr_paper_tdma/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:8080"
)

func main() {
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

	var base64_str string = "YWJj" //abc
	var dl_data []byte = []byte(base64_str)
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

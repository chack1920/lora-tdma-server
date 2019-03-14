package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/lioneie/lora-app-server/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:8080"
)

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
			MulticastGroupId: "4a21c7f8-4111-4e46-97c9-2986ca60bac5", //classC
			//MulticastGroupId: "2b38cd35-c36b-4b5d-91b1-5ab085d2335d",//classB
			FPort: 5,
			Data:  mc_data,
		},
	})
	if err != nil {
		log.Fatalf("send multicast error: %v", err)
	}
	log.Printf("enqueue multicast success, fcnt: %s", r.FCnt)
}
func main() {
	enqueue_multicastq()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	fmt.Println(<-sigChan, "signal received")
}

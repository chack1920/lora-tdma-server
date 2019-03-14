package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

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
}
func main() {
	sub_mqtt()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	fmt.Println(<-sigChan, "signal received")
}

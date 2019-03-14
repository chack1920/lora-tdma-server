package mqttpubsub

import (
	//"bytes"
	//"crypto/tls"
	//"crypto/x509"
	//"encoding/json"
	//"io/ioutil"
	"sync"
	//"text/template"
	//"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	//newMQTT "github.com/lioneie/lora-gateway-bridge/internal/backend/mqtt"
	//"github.com/lioneie/loraserver/api/gw"
	//"github.com/lioneie/lorawan"
	//"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Backend implements a MQTT pub-sub backend.
type Backend struct {
	conn  mqtt.Client
	mutex sync.RWMutex
}

// NewBackend creates a new Backend.
func NewBackend() (*Backend, error) {
	broker := "tcp://127.0.0.1:1883"

	b := Backend{}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	//opts.SetUsername(b.config.Auth.Generic.Username)
	//opts.SetPassword(b.config.Auth.Generic.Password)
	opts.SetCleanSession(true)
	opts.SetClientID("")
	opts.SetOnConnectHandler(b.onConnected)
	opts.SetConnectionLostHandler(b.onConnectionLost)

	maxReconnectInterval := 10 * time.Minute
	log.Infof("backend: set max reconnect interval: %s", maxReconnectInterval)
	opts.SetMaxReconnectInterval(maxReconnectInterval)

	log.WithField("server", broker).Info("backend: connecting to mqtt broker")
	b.conn = mqtt.NewClient(opts)

	if token := b.conn.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &b, nil
}

// Close closes the backend.
func (b *Backend) Close() {
	b.conn.Disconnect(250) // wait 250 milisec to complete pending actions
}

func (b *Backend) SubscribeAppTopic() error {
	var topic = "application/+/device/+/+"
	var qos byte = 0
	var callback = b.appPacketHandler
	if token := b.conn.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		log.Error("subscribe topic error: ", token.Error())
		return token.Error()
	}
	log.WithFields(log.Fields{
		"topic": topic,
		"qos":   qos,
	}).Info("backend: subscribing to topic")
	return nil
}

func (b *Backend) appPacketHandler(c mqtt.Client, msg mqtt.Message) {
	log.WithField("topic", msg.Topic()).Info("backend: app packet received")
}

func (b *Backend) onConnected(c mqtt.Client) {
	//mqttEventCounter("connected")

	defer b.mutex.RUnlock()
	b.mutex.RLock()

	log.Info("backend: connected to mqtt broker")
}

func (b *Backend) onConnectionLost(c mqtt.Client, reason error) {
	//mqttEventCounter("connection_lost")
	log.WithError(reason).Error("backend: mqtt connection error")
}

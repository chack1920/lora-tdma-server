package asclient

import (
	//"bytes"
	"context"
	//"crypto/tls"
	//"crypto/x509"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/credentials"

	pb "github.com/lioneie/lora-app-server/api"
)

type Pool interface {
	GetDeviceQueueServiceClient(hostname string) (pb.DeviceQueueServiceClient, error)
	GetMulticastGroupServiceClient(hostname string) (pb.MulticastGroupServiceClient, error)
}

type devClient struct {
	client     pb.DeviceQueueServiceClient
	clientConn *grpc.ClientConn
}

type multicastClient struct {
	client     pb.MulticastGroupServiceClient
	clientConn *grpc.ClientConn
}

type pool struct {
	sync.RWMutex
	devClients       map[string]devClient
	multicastClients map[string]multicastClient
}

// NewPool creates a Pool.
func NewPool() Pool {
	return &pool{
		devClients:       make(map[string]devClient),
		multicastClients: make(map[string]multicastClient),
	}
}

func (p *pool) GetDeviceQueueServiceClient(hostname string) (pb.DeviceQueueServiceClient, error) {
	return nil, nil //TODO
}

func (p *pool) GetMulticastGroupServiceClient(hostname string) (pb.MulticastGroupServiceClient, error) {
	defer p.Unlock()
	p.Lock()

	var connect bool
	c, ok := p.multicastClients[hostname]
	if !ok {
		connect = true
	}

	if connect {
		clientConn, err := p.createClientConn(hostname)
		if err != nil {
			return nil, errors.Wrap(err, "create device queue api client error")
		}
		client := pb.NewMulticastGroupServiceClient(clientConn)

		c = multicastClient{
			client:     client,
			clientConn: clientConn,
		}
		p.multicastClients[hostname] = c
	}

	return c.client, nil
}

func (p *pool) createClientConn(hostname string) (*grpc.ClientConn, error) {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	nsOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			grpc_logrus.UnaryClientInterceptor(logrusEntry, logrusOpts...),
		),
		grpc.WithStreamInterceptor(
			grpc_logrus.StreamClientInterceptor(logrusEntry, logrusOpts...),
		),
	}

	nsOpts = append(nsOpts, grpc.WithInsecure())
	log.WithField("server", hostname).Info("creating insecure client")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	clientConn, err := grpc.DialContext(ctx, hostname, nsOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "dial api error")
	}

	return clientConn, nil
}

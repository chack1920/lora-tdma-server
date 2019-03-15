package multicast

import (
	"context"

	//"github.com/gofrs/uuid"
	//"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/lioneie/lora-tdma-server/internal/config"
	//"github.com/lioneie/lora-app-server/internal/storage"
	pb "github.com/lioneie/lora-app-server/api"
	//"github.com/lioneie/lorawan"
)

// Enqueue adds the given payload to the multicast-group queue.
func Enqueue(multicastGroupID string, fPort uint32, data []byte) (uint32, error) {
	hostname := config.C.AppServer.Bind
	c, err := config.C.AppServer.Pool.GetMulticastGroupServiceClient(hostname)
	if err != nil {
		return 0, errors.Wrap(err, "get multicast client error")
	}

	r, err := c.Enqueue(context.Background(), &pb.EnqueueMulticastQueueItemRequest{
		MulticastQueueItem: &pb.MulticastQueueItem{
			MulticastGroupId: multicastGroupID,
			FPort:            fPort,
			Data:             data,
		},
	})
	if err != nil {
		return 0, errors.Wrap(err, "send multicast error")
	}
	return r.FCnt, nil
}

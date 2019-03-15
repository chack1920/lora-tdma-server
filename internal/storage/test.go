package storage

import (
	"time"

	//"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	//"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	//"github.com/lioneie/lorawan"
)

type TestItem struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	//ScheduleAt              time.Time      `db:"schedule_at"`
	//EmitAtTimeSinceGPSEpoch *time.Duration `db:"emit_at_time_since_gps_epoch"`
	//MulticastGroupID        uuid.UUID      `db:"multicast_group_id"`
	//GatewayID               lorawan.EUI64  `db:"gateway_id"`
	FCnt uint32 `db:"f_cnt"`
	//FPort                   uint8          `db:"f_port"`
	//FRMPayload              []byte         `db:"frm_payload"`
}

// CreateMulticastQueueItem adds the given item to the queue.
func CreateTestItem(db sqlx.Queryer, qi *TestItem) error {
	qi.CreatedAt = time.Now()

	err := sqlx.Get(db, &qi.ID, `
                insert into test (
                        created_at,
                        f_cnt
                ) values ($1, $2)
                returning
                        id
                `,
		qi.CreatedAt,
		qi.FCnt,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":    qi.ID,
		"f_cnt": qi.FCnt,
	}).Info("teset-item created")

	return nil
}

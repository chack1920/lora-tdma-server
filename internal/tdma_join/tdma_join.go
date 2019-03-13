package tdma_join

import (
	"github.com/lioneie/lorawan"
)

func HandleTdmaJoinRequest(pl lorawan.TdmaReqPayload) lorawan.TdmaAnsPayload {
	var mc_seq uint8 = 55

	var ans lorawan.TdmaAnsPayload = lorawan.TdmaAnsPayload{
		DevEUI: pl.DevEUI,
		McSeq:  mc_seq,
	}
	return ans
}

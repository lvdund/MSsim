package sender

import (
	context2 "github.com/lvdund/mssim/internal/control_test_engine/gnb/context"
	"github.com/lvdund/mssim/internal/control_test_engine/ue/context"

	log "github.com/sirupsen/logrus"
)

func SendToGnb(ue *context.UEContext, message []byte) {
	SendToGnbMsg(ue, context2.UEMessage{IsNas: true, Nas: message})
}

func SendToGnbMsg(ue *context.UEContext, message context2.UEMessage) {
	ue.Lock()
	gnbRx := ue.GetGnbRx()
	if gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		gnbRx <- message
	}
	ue.Unlock()
}

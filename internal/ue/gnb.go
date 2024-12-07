package context

import (
	log "github.com/sirupsen/logrus"
	gnbContext "mssim/internal/gnb/context"
)

func (ue *UeContext) handleGnbMsg(msg gnbContext.UEMessage) {
	if msg.IsNas {
		ue.HandleNas(msg.Nas)
	} else if msg.GNBPduSessions[0] != nil {
		// Setup PDU Session
		ue.setupGtpInterface(msg)
	} else if msg.GNBRx != nil && msg.GNBTx != nil && msg.GNBInboundChannel != nil {
		log.Info("[UE] gNodeB is telling us to use another gNodeB")
		previousGnbRx := ue.gnbRx
		ue.SetGnbInboundChannel(msg.GNBInboundChannel)
		ue.gnbRx = msg.GNBRx
		ue.gnbTx = msg.GNBTx
		previousGnbRx <- gnbContext.UEMessage{ConnectionClosed: true}
		close(previousGnbRx)
	} else {
		log.Error("[UE] Received unknown message from gNodeB", msg)
	}
}

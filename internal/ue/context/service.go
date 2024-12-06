package context

import (
	log "github.com/sirupsen/logrus"
	gnbContext "mssim/internal/gnb/context"
	"os"
	"os/signal"
	"sync"
)

func (ue *UEContext) Service(wg *sync.WaitGroup, ueMgrCh chan UeTesterMessage) {
	// starting communication with GNB and listen.
	ue.InitConn(ue.GetGnbInboundChannel())
	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	// Block until a signal is received.
	loop := true
	for loop {
		select {
		case msg, open := <-ue.gnbTx:
			if !open {
				log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with gNB was closed")
				ue.gnbTx = nil
				break
			}
			ue.handleGnbMsg(msg)
		case msg, open := <-ueMgrCh:
			if !open {
				log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with scenario was closed")
				loop = false
				break
			}
			//loop = ueMgrHandler(msg, ue)
			loop = ue.handleExternalTrigger(msg)
		case <-ue.GetDRX():
			ue.verifyPaging()
		}
	}
	ue.Terminate()
	wg.Done()
}

func (ue *UEContext) verifyPaging() {
	gnbTx := make(chan gnbContext.UEMessage, 1)

	ue.GetGnbInboundChannel() <- gnbContext.UEMessage{GNBTx: gnbTx, FetchPagedUEs: true}
	msg := <-gnbTx
	for _, pagedUE := range msg.PagedUEs {
		if ue.Get5gGuti() != nil && pagedUE.FiveGSTMSI != nil && [4]uint8(pagedUE.FiveGSTMSI.FiveGTMSI.Value) == ue.GetTMSI5G() {
			ue.handleExternalTrigger(UeTesterMessage{Type: ServiceRequestTrigger})
			return
		}
	}
}

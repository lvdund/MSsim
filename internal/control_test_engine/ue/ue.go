package ue

import (
	"os"
	"os/signal"
	"sync"

	"mssim/config"
	context2 "mssim/internal/control_test_engine/gnb/context"
	"mssim/internal/control_test_engine/procedures"
	"mssim/internal/control_test_engine/ue/context"
	serviceGtp "mssim/internal/control_test_engine/ue/gtp/service"
	"mssim/internal/control_test_engine/ue/scenario"

	log "github.com/sirupsen/logrus"
)

func NewUE(conf config.Config, id int, ueMgrChannel chan procedures.UeTesterMessage, gnbInboundChannel chan context2.UEMessage, wg *sync.WaitGroup, logFile string) chan scenario.ScenarioMessage {
	// new UE instance.
	ue := &context.UEContext{}
	scenarioChan := make(chan scenario.ScenarioMessage)

	// new UE context
	ue.NewRanUeContext(
		conf.Ue.Msin,
		conf.Ue.GetUESecurityCapability(),
		conf.Ue.Key,
		conf.Ue.Opc,
		"c9e8763286b5b9ffbdf56e1297d0887b",
		conf.Ue.Amf,
		conf.Ue.Sqn,
		conf.Ue.Hplmn.Mcc,
		conf.Ue.Hplmn.Mnc,
		conf.Ue.RoutingIndicator,
		conf.Ue.Dnn,
		int32(conf.Ue.Snssai.Sst),
		conf.Ue.Snssai.Sd,
		conf.Ue.TunnelMode,
		scenarioChan,
		gnbInboundChannel,
		id, logFile)

	go func() {
		// starting communication with GNB and listen.
		ue.InitConn(ue.GetGnbInboundChannel())
		sigStop := make(chan os.Signal, 1)
		signal.Notify(sigStop, os.Interrupt)

		// Block until a signal is received.
		loop := true
		for loop {
			select {
			case msg, open := <-ue.GetGnbTx():
				if !open {
					log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with gNB was closed")
					ue.SetGnbTx(nil)
					break
				}
				gnbMsgHandler(msg, ue)
			case msg, open := <-ueMgrChannel:
				if !open {
					log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with scenario was closed")
					loop = false
					break
				}
				//loop = ueMgrHandler(msg, ue)
				loop = ue.HandleExternalTrigger(msg)
			case <-ue.GetDRX():
				verifyPaging(ue)
			}
		}
		ue.Terminate()
		wg.Done()
	}()

	return scenarioChan
}

func gnbMsgHandler(msg context2.UEMessage, ue *context.UEContext) {
	if msg.IsNas {
		ue.HandleNas(msg.Nas)
	} else if msg.GNBPduSessions[0] != nil {
		// Setup PDU Session
		serviceGtp.SetupGtpInterface(ue, msg)
	} else if msg.GNBRx != nil && msg.GNBTx != nil && msg.GNBInboundChannel != nil {
		log.Info("[UE] gNodeB is telling us to use another gNodeB")
		previousGnbRx := ue.GetGnbRx()
		ue.SetGnbInboundChannel(msg.GNBInboundChannel)
		ue.SetGnbRx(msg.GNBRx)
		ue.SetGnbTx(msg.GNBTx)
		previousGnbRx <- context2.UEMessage{ConnectionClosed: true}
		close(previousGnbRx)
	} else {
		log.Error("[UE] Received unknown message from gNodeB", msg)
	}
}

func verifyPaging(ue *context.UEContext) {
	gnbTx := make(chan context2.UEMessage, 1)

	ue.GetGnbInboundChannel() <- context2.UEMessage{GNBTx: gnbTx, FetchPagedUEs: true}
	msg := <-gnbTx
	for _, pagedUE := range msg.PagedUEs {
		if ue.Get5gGuti() != nil && pagedUE.FiveGSTMSI != nil && [4]uint8(pagedUE.FiveGSTMSI.FiveGTMSI.Value) == ue.GetTMSI5G() {
			ue.HandleExternalTrigger(procedures.UeTesterMessage{Type: procedures.ServiceRequest})
			return
		}
	}
}

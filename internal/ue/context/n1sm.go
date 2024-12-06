package context

import (
	"github.com/reogac/nas"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

func (ue *UEContext) handlePduSessionEstablishmentAccept(msg *nas.PduSessionEstablishmentAccept) {
	log.Info("[UE][NAS] Receiving PDU Session Establishment Accept")

	if msg.GetPti() != 1 {
		log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PTI not the expected value")
	}
	if msg.SelectedPduSessionType != 1 {
		log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PDU Session Type not the expected value")
	}

	// update PDU Session information.
	pduSessionId := msg.GetSessionId()
	pduSession, err := ue.GetPduSession(pduSessionId)
	if err != nil {
		log.Error("[UE][NAS] Receiving PDU Session Establishment Accept about an unknown PDU Session, id: ", pduSessionId)
		return
	}

	// change the state of ue(SM)(PDU Session Active).
	pduSession.SetStateSM_PDU_SESSION_ACTIVE()

	// get UE IP
	if msg.PduAddress != nil {
		UeIp := msg.PduAddress
		pduSession.SetIp(UeIp.Content())
		log.Info("[UE][NAS] PDU address received: ", pduSession.GetIp())
	}

	// get QoS Rules
	QosRule := msg.AuthorizedQosRules
	log.Info("[UE][NAS] PDU session QoS RULES: ", string(QosRule.Bytes))

	// get DNN
	if msg.Dnn != nil {
		log.Info("[UE][NAS] PDU session DNN: ", msg.Dnn.String())
	}

	// get SNSSAI
	if msg.SNssai != nil {
		sst := msg.SNssai.Sst
		sd := msg.SNssai.GetSd()
		log.Info("[UE][NAS] PDU session NSSAI -- sst: ", sst, " sd: ", sd)
	}

}
func (ue *UEContext) handlePduSessionEstablishmentReject(msg *nas.PduSessionEstablishmentReject) {
	log.Error("[UE][NAS] Receiving PDU Session Establishment Reject")

	pduSessionId := msg.GetSessionId()

	log.Error("[UE][NAS] PDU Session Establishment Reject for PDU Session ID ", pduSessionId, ", 5GSM Cause: ", cause5GSMToString(uint8(msg.GsmCause)))

	// Per 5GSM state machine in TS 24.501 - 6.1.3.2.1., we re-try the setup until it's successful
	pduSession, err := ue.GetPduSession(pduSessionId)
	if err != nil {
		log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
		return
	}
	if pduSession.T3580Retries < 5 {
		// T3580 Timer
		go func() {
			// Exponential backoff
			time.Sleep(time.Duration(math.Pow(5, float64(pduSession.T3580Retries))) * time.Second)
			ue.triggerInitPduSessionRequestInner(pduSession)
			pduSession.T3580Retries++
		}()
	} else {
		log.Error("[UE][NAS] We re-tried five times to create PDU Session ", pduSessionId, ", Aborting.")
	}
}

func (ue *UEContext) handlePduSessionReleaseCommand(msg *nas.PduSessionReleaseCommand) {
	log.Info("[UE][NAS] Receiving PDU Session Release Command")

	pduSessionId := msg.GetSessionId()
	pduSession, err := ue.GetPduSession(pduSessionId)
	if pduSession == nil || err != nil {
		log.Error("[UE][NAS] Unable to delete PDU Session ", pduSessionId, " from UE ", ue.GetMsin(), " as the PDU Session was not found. Ignoring.")
		return
	}
	ue.DeletePduSession(pduSessionId)
	log.Info("[UE][NAS] Successfully released PDU Session ", pduSessionId, " from UE Context")
	ue.triggerInitPduSessionReleaseComplete(pduSession)
}

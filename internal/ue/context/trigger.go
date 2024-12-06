package context

import (
	"github.com/reogac/nas"
	log "github.com/sirupsen/logrus"
	gnbContext "mssim/internal/gnb/context"
	"os"
	"time"
)

type UeTesterMessageType uint8

const (
	RegistrationTrigger UeTesterMessageType = iota
	DeregistrationTrigger
	NewPDUSessionTrigger
	DestroyPDUSessionTrigger
	TerminateTrigger
	KillTrigger
	IdleTrigger
	ServiceRequestTrigger
)

type UeTesterMessage struct {
	Type  UeTesterMessageType
	Param uint8
	//GnbChan chan context.UEMessage
}

func (ue *UEContext) handleExternalTrigger(msg UeTesterMessage) bool {
	loop := true
	switch msg.Type {
	case RegistrationTrigger:
		ue.triggerInitRegistration()
	case DeregistrationTrigger:
		ue.triggerInitDeregistration()
	case NewPDUSessionTrigger:
		ue.triggerInitPduSessionRequest()
	case DestroyPDUSessionTrigger:
		pdu, err := ue.GetPduSession(msg.Param)
		if err != nil {
			log.Error("[UE] Cannot release unknown PDU Session ID ", msg.Param)
			return loop
		}
		ue.triggerInitPduSessionRelease(pdu)
	case IdleTrigger:
		// We switch UE to IDLE
		ue.SetStateMM_IDLE()
		//trigger.SwitchToIdle(ue)
		ue.CreateDRX(25 * time.Millisecond)
	case ServiceRequestTrigger:
		if ue.GetStateMM() == MM5G_IDLE {
			ue.StopDRX()

			// Since gNodeB stopped communication after switching to Idle, we need to connect back to gNodeB
			ue.InitConn(ue.GetGnbInboundChannel())
			if ue.Get5gGuti() != nil {
				ue.triggerInitServiceRequest()
			} else {
				// If AMF did not assign us a GUTI, we have to fallback to the usual Registration/Authentification process
				// PDU Sessions will still be recovered
				//trigger.InitRegistration(ue)
			}
		}
	case TerminateTrigger:
		log.Info("[UE] Terminating UE as requested")
		// If UE is registered
		if len(ue.ExpFile) > 0 {
			if ExpFile, err := os.OpenFile(ue.ExpFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
				log.Errorf("Failed to create logfile " + ue.ExpFile)
			} else {
				LogExpResults(ExpFile, ue)
				ExpFile.Close()
			}
		}
		if ue.GetStateMM() == MM5G_REGISTERED {
			// Release PDU Sessions
			for i := uint8(1); i <= 16; i++ {
				pduSession, _ := ue.GetPduSession(i)
				if pduSession != nil {
					ue.triggerInitPduSessionRelease(pduSession)
					select {
					case <-pduSession.Wait:
					case <-time.After(500 * time.Millisecond):
						// If still unregistered after 500 ms, continue
					}
				}
			}
			// Initiate Deregistration
			ue.triggerInitDeregistration()
		}
		// Else, nothing to do
		loop = false
	case KillTrigger:
		loop = false
	}
	return loop
}

func (ue *UEContext) triggerInitRegistration() {
	log.Info("[UE] Initiating Registration")

	msg := &nas.RegistrationRequest{
		//TODO: build  RegistrationRequest content
	}
	nasCtx := ue.getNasContext()
	if nasCtx != nil {
		msg.SetSecurityHeader(nas.NasSecBoth)
	} else {
		msg.SetSecurityHeader(nas.NasSecNone)
	}
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatalf("[UE][NAS] Unable to encode Registration Request: %s", err)
	} else {
		// send to GNB.
		ue.SendNas(nasPdu)

		// change the state of ue for deregistered
		ue.SetStateMM_DEREGISTERED()
	}
}

func (ue *UEContext) triggerInitPduSessionRequest() {
	log.Info("[UE] Initiating New PDU Session")

	pduSession, err := ue.CreatePDUSession()
	if err != nil {
		log.Fatal("[UE][NAS] ", err)
		return
	}
	pduSession.Exp.CreatedTime = time.Now()
	ue.triggerInitPduSessionRequestInner(pduSession)
	pduSession.Exp.ActivatedTime = time.Now()
}

func (ue *UEContext) triggerInitPduSessionRequestInner(pduSession *UEPDUSession) {
	//TODO:build N1Sm pdu
	msg := &nas.UlNasTransport{
		//TODO: build content
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	} else {
		// change the state of ue(SM).
		pduSession.SetStateSM_PDU_SESSION_PENDING()

		// sending to GNB
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerInitPduSessionRelease(pduSession *UEPDUSession) {
	log.Info("[UE] Initiating Release of PDU Session ", pduSession.Id)

	if pduSession.GetStateSM() != SM5G_PDU_SESSION_ACTIVE {
		log.Warn("[UE][NAS] Skipping releasing the PDU Session ID ", pduSession.Id, " as it's not active")
		return
	}
	//TODO:build N1Sm pdu

	msg := &nas.UlNasTransport{
		//TODO: build content
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	} else {

		// change the state of ue(SM).
		pduSession.SetStateSM_PDU_SESSION_INACTIVE()

		// sending to GNB
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerInitPduSessionReleaseComplete(pduSession *UEPDUSession) {
	log.Info("[UE] Initiating PDU Session Release Complete for PDU Session", pduSession.Id)

	if pduSession.GetStateSM() != SM5G_PDU_SESSION_INACTIVE {
		log.Warn("[UE][NAS] Unable to send PDU Session Release Complete for a PDU Session which is not inactive")
		return
	}
	//TODO:build N1Sm pdu

	msg := &nas.UlNasTransport{
		//TODO: build content
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding ul nas transport and pdu session establishment request: ", err)
	} else {

		// sending to GNB
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerInitDeregistration() {
	log.Info("[UE] Initiating Deregistration")

	msg := &nas.DeregistrationRequestFromUe{
		//TODO: set content
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding deregistration request: ", err)
	} else {
		// send to GNB.
		ue.SendNas(nasPdu)
		// change the state of ue for deregistered
		ue.SetStateMM_DEREGISTERED()
	}
}

func (ue *UEContext) triggerInitIdentifyResponse() {
	log.Info("[UE] Initiating Identify Response")

	msg := &nas.IdentityResponse{
		//TODO: set content
	}
	nasCtx := ue.getNasContext()
	if nasCtx != nil {
		msg.SetSecurityHeader(nas.NasSecBoth)
	} else {
		msg.SetSecurityHeader(nas.NasSecNone)
	}

	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding identity request: ", err)
	} else {
		// send to GNB.
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerInitConfigurationUpdateComplete() {
	log.Info("[UE] Initiating Configuration Update Complete")

	msg := &nas.ConfigurationUpdateComplete{
		//TODO: set content
	}
	nasCtx := ue.getNasContext()
	msg.SetSecurityHeader(nas.NasSecBoth)

	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding Configuration Update Complete: ", err)
	} else {
		// send to GNB.
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerInitServiceRequest() {
	log.Info("[UE] Initiating Service Request")

	msg := &nas.ServiceRequest{
		//TODO: set content
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)

	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatalf("Error encoding Service Request Msg", err)
	} else {

		// send to GNB.
		ue.SendNas(nasPdu)
	}
}

func (ue *UEContext) triggerSwitchToIdle() {
	log.Info("[UE] Switching to 5GMM-IDLE")

	// send to GNB.
	ue.SendGnb(gnbContext.UEMessage{Idle: true})
}

func (ue *UEContext) InitConn(gnbInboundChannel chan gnbContext.UEMessage) {
	ue.gnbRx = make(chan gnbContext.UEMessage, 1)
	ue.gnbTx = make(chan gnbContext.UEMessage, 1)

	// Send channels to gNB
	gnbInboundChannel <- gnbContext.UEMessage{GNBTx: ue.gnbTx, GNBRx: ue.gnbRx, PrUeId: ue.GetPrUeId(), Tmsi: ue.Get5gGuti()}
	msg := <-ue.gnbTx
	ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
}

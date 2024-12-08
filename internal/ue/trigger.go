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

func (ue *UeContext) handleExternalTrigger(msg UeTesterMessage) bool {
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
			ue.initConn(ue.GetGnbInboundChannel())
			if ue.guti != nil {
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

func (ue *UeContext) sendN1Sm(n1Sm []byte) {
	msg := &nas.UlNasTransport{
		PayloadContainer:     nas.Bytes(n1Sm),
		PayloadContainerType: nas.Uint8(nas.PayloadContainerTypeN1SMInfo),
	}

	if len(ue.Dnn) > 0 {
		msg.Dnn = nas.NewDnn(ue.Dnn)
	}

	msg.SNssai.Set(uint8(ue.Snssai.Sst), ue.Snssai.Sd)

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error sending ul nas transport and pdu session establishment request: ", err)
	} else {

		// sending to GNB
		ue.sendNas(nasPdu)
	}

}

func (ue *UeContext) triggerInitPduSessionRequest() {
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

func (ue *UeContext) triggerInitPduSessionRequestInner(pduSession *UEPDUSession) {
	n1Sm := new(nas.PduSessionEstablishmentRequest)
	n1Sm.PduSessionType = new(nas.Uint8)
	*n1Sm.PduSessionType = nas.Uint8(nas.PduSessionTypeIpv4)
	n1Sm.IntegrityProtectionMaximumDataRate = nas.NewIntegrityProtectionMaximumDataRate(0xff, 0xff)
	/*
		//TODO: add Extended PCOs: refer to SMF (handle
		//PduSessionEstablishmentRequest)
				msg.ExtendedProtocolConfigurationOptions = nasType.NewExtendedProtocolConfigurationOptions(nasMessage.PDUSessionEstablishmentRequestExtendedProtocolConfigurationOptionsType)
				protocolConfigurationOptions := nasConvert.NewProtocolConfigurationOptions()
				protocolConfigurationOptions.AddIPAddressAllocationViaNASSignallingUL()
				protocolConfigurationOptions.AddDNSServerIPv4AddressRequest()
				protocolConfigurationOptions.AddDNSServerIPv6AddressRequest()
				pcoContents := protocolConfigurationOptions.Marshal()
				pcoContentsLength := len(pcoContents)
				msg.ExtendedProtocolConfigurationOptions.SetLen(uint16(pcoContentsLength))
				msg.ExtendedProtocolConfigurationOptions.SetExtendedProtocolConfigurationOptionsContents(pcoContents)
	*/

	n1Sm.SetPti(1)
	n1Sm.SetSessionId(pduSession.Id)

	n1SmPdu, _ := nas.EncodeSm(n1Sm)
	ue.sendN1Sm(n1SmPdu)

	// change the state of ue(SM).
	pduSession.SetStateSM_PDU_SESSION_PENDING()
}

func (ue *UeContext) triggerInitPduSessionRelease(pduSession *UEPDUSession) {
	log.Info("[UE] Initiating Release of PDU Session ", pduSession.Id)

	if pduSession.GetStateSM() != SM5G_PDU_SESSION_ACTIVE {
		log.Warn("[UE][NAS] Skipping releasing the PDU Session ID ", pduSession.Id, " as it's not active")
		return
	}
	n1Sm := new(nas.PduSessionReleaseRequest)

	n1Sm.SetPti(1)
	n1Sm.SetSessionId(pduSession.Id)
	n1SmPdu, _ := nas.EncodeSm(n1Sm)
	ue.sendN1Sm(n1SmPdu)
	// change the state of ue(SM).
	pduSession.SetStateSM_PDU_SESSION_INACTIVE()
}

func (ue *UeContext) triggerInitPduSessionReleaseComplete(pduSession *UEPDUSession) {
	log.Info("[UE] Initiating PDU Session Release Complete for PDU Session", pduSession.Id)

	if pduSession.GetStateSM() != SM5G_PDU_SESSION_INACTIVE {
		log.Warn("[UE][NAS] Unable to send PDU Session Release Complete for a PDU Session which is not inactive")
		return
	}
	n1Sm := new(nas.PduSessionReleaseComplete)

	n1Sm.SetPti(1) //must be same as received command message
	n1Sm.SetSessionId(pduSession.Id)
	n1SmPdu, _ := nas.EncodeSm(n1Sm)
	ue.sendN1Sm(n1SmPdu)
}

func (ue *UeContext) triggerInitRegistration() {
	log.Info("[UE] Initiating Registration")

	msg := &nas.RegistrationRequest{
		UeSecurityCapability: ue.secCap,
	}
	msg.RegistrationType.Value = nas.RegistrationType5GSInitialRegistration

	if ue.guti != nil {
		msg.MobileIdentity = nas.MobileIdentity{
			Id: ue.guti,
		}
	} else {
		msg.MobileIdentity = ue.suci
	}
	if ue.secCtx != nil {
		msg.Ngksi.Id = ue.secCtx.NgKsi().Id
	} else {
		msg.Ngksi.Id = 7
	}
	var gmmCap [13]byte
	gmmCap[0] = 0x07
	msg.GmmCapability = new(nas.GmmCapability)
	msg.GmmCapability.Bytes = nas.Bytes(gmmCap[:])

	var pduFlag [16]bool
	hasPdu := false
	for i, pduSession := range ue.PduSession {
		if pduSession != nil {
			hasPdu = true
			pduFlag[i] = true
		}
	}

	if hasPdu {
		msg.UplinkDataStatus = new(nas.UplinkDataStatus)
		msg.UplinkDataStatus.Set(pduFlag)

		msg.PduSessionStatus = new(nas.PduSessionStatus)
		msg.PduSessionStatus.Set(pduFlag)
	}
	msg.SetSecurityHeader(nas.NasSecNone)

	//TODO: RequestedNssai
	//
	nasPdu, _ := nas.EncodeMm(nil, msg)
	//Keep a copy of this registration request
	ue.nasPdu = make([]byte, len(nasPdu))
	copy(ue.nasPdu, nasPdu)

	if hasPdu {
		//encrypt the request
		nasCtx := ue.getNasContext()                   //must be non-nil
		cipher, _ := nasCtx.EncryptMmContainer(nasPdu) //ignore error for now
		//embed the encrypted request into the original one
		msg.NasMessageContainer = new(nas.Bytes)
		*msg.NasMessageContainer = cipher
		//reset UplinkDataStatus and PduSessionStatus
		msg.UplinkDataStatus = nil
		msg.PduSessionStatus = nil
		//now plaintext-encode again
		nasPdu, _ = nas.EncodeMm(nil, msg)
	}
	// send to GNB.
	ue.sendNas(nasPdu)

	// change the state of ue for deregistered
	ue.SetStateMM_DEREGISTERED()
}

func (ue *UeContext) triggerInitDeregistration() {
	log.Info("[UE] Initiating Deregistration")

	msg := &nas.DeregistrationRequestFromUe{
		Ngksi: *ue.secCtx.NgKsi(),
	}
	msg.DeRegistrationType.SetSwitchOff(true)
	msg.DeRegistrationType.SetReregistration(false)
	msg.DeRegistrationType.SetAccessType(nas.AccessType3GPP)
	if ue.guti != nil {
		msg.MobileIdentity.Id = ue.guti
	} else {
		msg.MobileIdentity = ue.suci
	}

	nasCtx := ue.getNasContext() //must be non nil
	msg.SetSecurityHeader(nas.NasSecBoth)
	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding deregistration request: ", err)
	} else {
		// send to GNB.
		ue.sendNas(nasPdu)
		// change the state of ue for deregistered
		ue.SetStateMM_DEREGISTERED()
	}
}

func (ue *UeContext) triggerInitIdentifyResponse() {
	log.Info("[UE] Initiating Identify Response")

	msg := &nas.IdentityResponse{
		MobileIdentity: ue.suci, //TODO: can be SUCI/IMEISV etc
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
		ue.sendNas(nasPdu)
	}
}

func (ue *UeContext) triggerInitConfigurationUpdateComplete() {
	log.Info("[UE] Initiating Configuration Update Complete")

	msg := &nas.ConfigurationUpdateComplete{}
	nasCtx := ue.getNasContext()
	msg.SetSecurityHeader(nas.NasSecBoth)

	if nasPdu, err := nas.EncodeMm(nasCtx, msg); err != nil {
		log.Fatal("[UE][NAS] Error encoding Configuration Update Complete: ", err)
	} else {
		// send to GNB.
		ue.sendNas(nasPdu)
	}
}

func (ue *UeContext) triggerInitServiceRequest() {
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
		ue.sendNas(nasPdu)
	}
}

func (ue *UeContext) triggerSwitchToIdle() {
	log.Info("[UE] Switching to 5GMM-IDLE")

	// send to GNB.
	ue.sendGnb(gnbContext.UEMessage{Idle: true})
}

func (ue *UeContext) initConn(gnbInboundChannel chan gnbContext.UEMessage) {
	ue.gnbRx = make(chan gnbContext.UEMessage, 1)
	ue.gnbTx = make(chan gnbContext.UEMessage, 1)

	// Send channels to gNB
	gnbInboundChannel <- gnbContext.UEMessage{GNBTx: ue.gnbTx, GNBRx: ue.gnbRx, PrUeId: int64(ue.id), Tmsi: ue.guti}
	msg := <-ue.gnbTx
	ue.snn = deriveSNN(msg.Mcc, msg.Mnc)
}

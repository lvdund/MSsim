package context

import (
	"github.com/reogac/nas"
	"github.com/reogac/sbi/models"
	log "github.com/sirupsen/logrus"
	gnbContext "mssim/internal/gnb/context"
)

func (ue *UeContext) SendGnb(message gnbContext.UEMessage) {
	ue.Lock()
	if ue.gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		ue.gnbRx <- message
	}
	ue.Unlock()
}

func (ue *UeContext) SendNas(nasPdu []byte) {
	ue.SendGnb(gnbContext.UEMessage{IsNas: true, Nas: nasPdu})
}

func (ue *UeContext) HandleNas(nasPdu []byte) {
	// check if nasPdu is empty
	if len(nasPdu) == 0 {
		log.Errorf("[UE][NAS] NAS message is empty")
		return
	}

	//decode nas message with current security context
	var nasMsg nas.NasMessage
	var err error
	if nasMsg, err = nas.Decode(ue.secCtx.NasContext(), nasPdu); err != nil {
		log.Errorf("Decode NasUl failed: %s", err.Error())
		return
	}
	//is N1Mm existed?
	gmm := nasMsg.Gmm
	if gmm == nil {
		log.Errorf("[UE][NAS] NAS message is has no N1MM content")
		return
	}

	switch gmm.MsgType {
	case nas.AuthenticationRequestMsgType:
		// handle authentication request.
		log.Info("[UE][NAS] Receive Authentication Request")
		ue.handleAuthenticationRequest(gmm.AuthenticationRequest)

	case nas.AuthenticationRejectMsgType:
		// handle authentication reject.
		log.Info("[UE][NAS] Receive Authentication Reject")
		ue.handleAuthenticationReject(gmm.AuthenticationReject)

	case nas.IdentityRequestMsgType:
		log.Info("[UE][NAS] Receive Identify Request")
		// handle identity request.
		ue.handleIdentityRequest(gmm.IdentityRequest)

	case nas.SecurityModeCommandMsgType:
		// handle security mode command.
		log.Info("[UE][NAS] Receive Security Mode Command")
		/*
			if !newSecurityContext {
				log.Warn("Received Security Mode Command with security header different from \"Integrity protected with new 5G NAS security context\" ")
			}
		*/
		ue.handleSecurityModeCommand(gmm.SecurityModeCommand)

	case nas.RegistrationAcceptMsgType:
		// handle registration accept.
		log.Info("[UE][NAS] Receive Registration Accept")
		ue.handleRegistrationAccept(gmm.RegistrationAccept)

	case nas.ConfigurationUpdateCommandMsgType:
		log.Info("[UE][NAS] Receive Configuration Update Command")
		ue.handleConfigurationUpdateCommand(gmm.ConfigurationUpdateCommand)

	case nas.DlNasTransportMsgType:
		// handle DL NAS Transport.
		log.Info("[UE][NAS] Receive DL NAS Transport")
		handleCause5GMM(gmm.DlNasTransport.GmmCause)
		ue.handleDlNasTransport(gmm.DlNasTransport)

	case nas.ServiceAcceptMsgType:
		// handle service reject
		log.Info("[UE][NAS] Receive Service Accept")
		ue.handleServiceAccept(gmm.ServiceAccept)

	case nas.ServiceRejectMsgType:
		// handle service reject
		log.Error("[UE][NAS] Receive Service Reject")
		handleCause5GMM(&gmm.ServiceReject.GmmCause)

	case nas.RegistrationRejectMsgType:
		// handle registration reject
		log.Error("[UE][NAS] Receive Registration Reject")
		handleCause5GMM(&gmm.RegistrationReject.GmmCause)

	case nas.GmmStatusMsgType:
		log.Error("[UE][NAS] Receive Status 5GMM")
		handleCause5GMM(&gmm.GmmStatus.GmmCause)

		//	case nas.GsmStatusMsgType:
		//		log.Error("[UE][NAS] Receive Status 5GSM")
		//		handleCause5GSM(&gmm.GsmStatus.GsmCause)

	default:
		log.Warnf("[UE][NAS] Received unknown NAS message 0x%x", nasMsg.Gmm.MsgType)
	}

}

func handleCause5GSM(cause *nas.Uint8) {
	if cause != nil {
		log.Error("[UE][NAS] UE received a 5GSM Failure, cause: ", cause5GSMToString(uint8(*cause)))
	}
}

func handleCause5GMM(cause *nas.Uint8) {
	if cause != nil {
		log.Error("[UE][NAS] UE received a 5GMM Failure, cause: ", cause5GMMToString(uint8(*cause)))
	}
}

func (ue *UeContext) handleAuthenticationReject(message *nas.AuthenticationReject) {
	log.Info("[UE][NAS] Authentication of UE ", ue.GetUeId(), " failed")
	ue.SetStateMM_DEREGISTERED()
}

func (ue *UeContext) handleAuthenticationRequest(message *nas.AuthenticationRequest) {
	var responsePdu []byte
	var response nas.GmmMessage

	if message.Ngksi.Id == 7 {
		log.Fatal("[UE][NAS] Error in Authentication Request, ngKSI not the expected value")
	}

	if len(message.Abba) == 0 {
		log.Fatal("[UE][NAS] Error in Authentication Request, ABBA Content is empty")
	}
	if message.AuthenticationParameterRand == nil {
		log.Fatal("[UE][NAS] Error in Authentication Request, RAND is missing")
	}

	if message.AuthenticationParameterAutn == nil {
		log.Fatal("[UE][NAS] Error in Authentication Request, AUTN is missing")
	}
	// getting RAND and AUTN from the message.
	rand := []byte(*message.AuthenticationParameterRand)
	autn := []byte(*message.AuthenticationParameterAutn)

	// getting resStar
	paramAutn, check := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, rand[:], ue.UeSecurity.Snn, autn[:])
	switch check {

	case "MAC failure":
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: FAILED")
		log.Info("[UE][NAS] Send authentication failure with MAC failure")
		response = &nas.AuthenticationFailure{
			GmmCause: nas.Uint8(nas.Cause5GMMMACFailure),
		}
		// not change the state of UE.
	case "SQN failure":
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message: INVALID")
		log.Info("[UE][NAS] Send authentication failure with Synch failure")
		msg := &nas.AuthenticationFailure{
			GmmCause:                       nas.Uint8(nas.Cause5GMMSynchFailure),
			AuthenticationFailureParameter: new(nas.Bytes),
		}
		*msg.AuthenticationFailureParameter = paramAutn
		response = msg
		// not change the state of UE.

	case "successful":
		// getting NAS Authentication Response.
		log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
		log.Info("[UE][NAS][SQN] SQN of the authentication request message: VALID")
		log.Info("[UE][NAS] Send authentication response")
		//authenticationResponse = mm_5gs.AuthenticationResponse(paramAutn, "")
		msg := &nas.AuthenticationResponse{
			AuthenticationResponseParameter: new(nas.Bytes),
		}
		*msg.AuthenticationResponseParameter = paramAutn
		response = msg
		// change state of UE for registered-initiated
		ue.SetStateMM_REGISTERED_INITIATED()
	}
	responsePdu, _ = nas.EncodeMm(nil, response)
	// sending to GNB
	ue.SendNas(responsePdu)
}

func (ue *UeContext) handleSecurityModeCommand(message *nas.SecurityModeCommand) {
	if message.Ngksi.Id == 7 {
		log.Fatal("[UE][NAS] Error in Security Mode Command, ngKSI not the expected value")
	}

	switch ue.UeSecurity.CipheringAlg {
	case 0:
		log.Info("[UE][NAS] Type of ciphering algorithm is 5G-EA0")
	case 1:
		log.Info("[UE][NAS] Type of ciphering algorithm is 128-5G-EA1")
	case 2:
		log.Info("[UE][NAS] Type of ciphering algorithm is 128-5G-EA2")
	}

	switch ue.UeSecurity.IntegrityAlg {
	case 0:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 5G-IA0")
	case 1:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 128-5G-IA1")
	case 2:
		log.Info("[UE][NAS] Type of integrity protection algorithm is 128-5G-IA2")
	}

	rinmr := false
	if message.AdditionalSecurityInformation != nil {
		// checking BIT RINMR that triggered registration request in security mode complete.
		rinmr = message.AdditionalSecurityInformation.GetRetransmission()
	}

	ue.UeSecurity.NgKsi.Ksi = int(message.Ngksi.Id)

	// NgKsi: TS 24.501 9.11.3.32
	switch message.Ngksi.Tsc {
	case nas.TscNative:
		ue.UeSecurity.NgKsi.Tsc = models.SCTYPE_NATIVE
	case nas.TscMapped:
		ue.UeSecurity.NgKsi.Tsc = models.SCTYPE_MAPPED
	}
	//TODO: activate security context

	imeisv := new(nas.Imei)
	imeisv.Parse("1111111111111111") //dummy imei
	response := &nas.SecurityModeComplete{
		Imeisv: &nas.MobileIdentity{
			Id: imeisv,
		},
	}
	if rinmr {
		nasCtx := ue.getNasContext() //must be non-nil
		//encrypt last sending registration request
		cipher, _ := nasCtx.EncryptMmContainer(ue.nasPdu)
		response.NasMessageContainer = new(nas.Bytes)
		*response.NasMessageContainer = cipher
	}

	nasCtx := ue.getNasContext()
	responsePdu, _ := nas.EncodeMm(nasCtx, response)
	// sending to GNB
	ue.SendNas(responsePdu)
}

func (ue *UeContext) handleRegistrationAccept(message *nas.RegistrationAccept) {

	// change the state of ue for registered
	ue.SetStateMM_REGISTERED()

	// saved 5g GUTI and others information.
	if message.Guti != nil {
		ue.Set5gGuti(message.Guti)
	} else {
		log.Warn("[UE][NAS] UE was not assigned a 5G-GUTI by AMF")
	}

	// use the slice allowed by the network
	// in PDU session request
	if ue.Snssai.Sst == 0 && message.AllowedNssai != nil {
		// check the allowed NSSAI received from the 5GC
		snssai := message.AllowedNssai.List[0] //very sloppy, need checking

		// update UE slice selected for PDU Session
		ue.Snssai.Sst = int(snssai.Sst)
		ue.Snssai.Sd = snssai.GetSd()

		log.Warn("[UE][NAS] ALLOWED NSSAI: SST: ", ue.Snssai.Sst, " SD: ", ue.Snssai.Sd)
	}

	log.Info("[UE][NAS] UE 5G GUTI: ", ue.Get5gGuti())

	// getting NAS registration complete.
	response := &nas.RegistrationComplete{}
	//TODO: set SORTransparentContainer if needed

	nasCtx := ue.getNasContext() //must be non-nil
	responsePdu, _ := nas.EncodeMm(nasCtx, response)
	// sending to GNB
	ue.SendNas(responsePdu)
}

func (ue *UeContext) handleServiceAccept(message *nas.ServiceAccept) {
	// change the state of ue for registered
	ue.SetStateMM_REGISTERED()
}

func (ue *UeContext) handleDlNasTransport(message *nas.DlNasTransport) {

	if uint8(message.PayloadContainerType) != nas.PayloadContainerTypeN1SMInfo {
		log.Fatal("[UE][NAS] Error in DL NAS Transport, Payload Container Type not expected value")
	}

	if message.PduSessionId == nil {
		log.Fatal("[UE][NAS] Error in DL NAS Transport, PDU Session ID is missing")
	}

	//decode N1Sm message
	nasMsg, err := nas.Decode(nil, message.PayloadContainer)

	if err != nil {
		log.Fatal("[UE][NAS] Error in DL NAS Transport, fail to decode N1Sm")
	}
	gsm := nasMsg.Gsm
	if gsm == nil {
		log.Fatal("[UE][NAS] Error in DL NAS Transport, N1Sm is missing")
	}

	switch gsm.MsgType {
	case nas.PduSessionEstablishmentAcceptMsgType:
		ue.handlePduSessionEstablishmentAccept(gsm.PduSessionEstablishmentAccept)

	case nas.PduSessionReleaseCommandMsgType:
		ue.handlePduSessionReleaseCommand(gsm.PduSessionReleaseCommand)
	case nas.PduSessionEstablishmentRejectMsgType:
		ue.handlePduSessionEstablishmentReject(gsm.PduSessionEstablishmentReject)

	default:
		log.Error("[UE][NAS] Receiving Unknown Dl NAS Transport message!! ", gsm.MsgType)
	}
}

func (ue *UeContext) handleIdentityRequest(message *nas.IdentityRequest) {

	switch uint8(message.IdentityType) {
	case nas.MobileIdentity5GSTypeSuci:
		log.Info("[UE][NAS] Requested SUCI 5GS type")
	default:
		log.Fatal("[UE][NAS] Only SUCI identity is supported for now inside MSsim")
	}

	ue.triggerInitIdentifyResponse()
}

func (ue *UeContext) handleConfigurationUpdateCommand(message *nas.ConfigurationUpdateCommand) {
	ue.triggerInitConfigurationUpdateComplete()
}

package context

import (
	"github.com/reogac/nas"
	log "github.com/sirupsen/logrus"
	gnbContext "mssim/internal/gnb/context"
)

func (ue *UEContext) SendGnb(message gnbContext.UEMessage) {
	ue.Lock()
	if ue.gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		ue.gnbRx <- message
	}
	ue.Unlock()
}

func (ue *UEContext) SendNas(nasPdu []byte) {
	ue.SendGnb(gnbContext.UEMessage{IsNas: true, Nas: nasPdu})
}

func (ue *UEContext) HandleNas(nasPdu []byte) {
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

func (ue *UEContext) handleAuthenticationReject(message *nas.AuthenticationReject) {
	/*
		log.Info("[UE][NAS] Authentication of UE ", ue.GetUeId(), " failed")

		ue.SetStateMM_DEREGISTERED()
	*/
}

func (ue *UEContext) handleAuthenticationRequest(message *nas.AuthenticationRequest) {
	/*
		var authenticationResponse []byte

		// check the mandatory fields
		if reflect.ValueOf(message.AuthenticationRequest.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in Authentication Request, Extended Protocol is missing")
		}

		if message.AuthenticationRequest.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in Authentication Request, Extended Protocol not the expected value")
		}

		if message.AuthenticationRequest.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Authentication Request, Spare Half Octet not the expected value")
		}

		if message.AuthenticationRequest.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in Authentication Request, Security Header Type not the expected value")
		}

		if reflect.ValueOf(message.AuthenticationRequest.AuthenticationRequestMessageIdentity).IsZero() {
			log.Fatal("[UE][NAS] Error in Authentication Request, Message Type is missing")
		}

		if message.AuthenticationRequest.AuthenticationRequestMessageIdentity.GetMessageType() != 86 {
			log.Fatal("[UE][NAS] Error in Authentication Request, Message Type not the expected value")
		}

		if message.AuthenticationRequest.SpareHalfOctetAndNgksi.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Authentication Request, Spare Half Octet not the expected value")
		}

		if message.AuthenticationRequest.SpareHalfOctetAndNgksi.GetNasKeySetIdentifiler() == 7 {
			log.Fatal("[UE][NAS] Error in Authentication Request, ngKSI not the expected value")
		}

		if reflect.ValueOf(message.AuthenticationRequest.ABBA).IsZero() {
			log.Fatal("[UE][NAS] Error in Authentication Request, ABBA is missing")
		}

		if message.AuthenticationRequest.GetABBAContents() == nil {
			log.Fatal("[UE][NAS] Error in Authentication Request, ABBA Content is missing")
		}

		// getting RAND and AUTN from the message.
		rand := message.AuthenticationRequest.GetRANDValue()
		autn := message.AuthenticationRequest.GetAUTN()

		// getting resStar
		paramAutn, check := ue.DeriveRESstarAndSetKey(ue.UeSecurity.AuthenticationSubs, rand[:], ue.UeSecurity.Snn, autn[:])

		switch check {

		case "MAC failure":
			log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: FAILED")
			log.Info("[UE][NAS] Send authentication failure with MAC failure")
			authenticationResponse = mm_5gs.AuthenticationFailure("MAC failure", "", paramAutn)
			// not change the state of UE.

		case "SQN failure":
			log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
			log.Info("[UE][NAS][SQN] SQN of the authentication request message: INVALID")
			log.Info("[UE][NAS] Send authentication failure with Synch failure")
			authenticationResponse = mm_5gs.AuthenticationFailure("SQN failure", "", paramAutn)
			// not change the state of UE.

		case "successful":
			// getting NAS Authentication Response.
			log.Info("[UE][NAS][MAC] Authenticity of the authentication request message: OK")
			log.Info("[UE][NAS][SQN] SQN of the authentication request message: VALID")
			log.Info("[UE][NAS] Send authentication response")
			authenticationResponse = mm_5gs.AuthenticationResponse(paramAutn, "")

			// change state of UE for registered-initiated
			ue.SetStateMM_REGISTERED_INITIATED()
		}

		// sending to GNB
		sender.SendToGnb(ue, authenticationResponse)
	*/
}

func (ue *UEContext) handleSecurityModeCommand(message *nas.SecurityModeCommand) { // check the mandatory fields
	/*
		if reflect.ValueOf(message.SecurityModeCommand.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Extended Protocol is missing")
		}

		if message.SecurityModeCommand.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Extended Protocol not the expected value")
		}

		if message.SecurityModeCommand.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Security Header Type not the expected value")
		}

		if message.SecurityModeCommand.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Spare Half Octet not the expected value")
		}

		if reflect.ValueOf(message.SecurityModeCommand.SecurityModeCommandMessageIdentity).IsZero() {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Message Type is missing")
		}

		if message.SecurityModeCommand.SecurityModeCommandMessageIdentity.GetMessageType() != 93 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Message Type not the expected value")
		}

		if reflect.ValueOf(message.SecurityModeCommand.SelectedNASSecurityAlgorithms).IsZero() {
			log.Fatal("[UE][NAS] Error in Security Mode Command, NAS Security Algorithms is missing")
		}

		if message.SecurityModeCommand.SpareHalfOctetAndNgksi.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Spare Half Octet is missing")
		}

		if message.SecurityModeCommand.SpareHalfOctetAndNgksi.GetNasKeySetIdentifiler() == 7 {
			log.Fatal("[UE][NAS] Error in Security Mode Command, ngKSI not the expected value")
		}

		if reflect.ValueOf(message.SecurityModeCommand.ReplayedUESecurityCapabilities).IsZero() {
			log.Fatal("[UE][NAS] Error in Security Mode Command, Replayed UE Security Capabilities is missing")
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

		rinmr := uint8(0)
		if message.SecurityModeCommand.Additional5GSecurityInformation != nil {
			// checking BIT RINMR that triggered registration request in security mode complete.
			rinmr = message.SecurityModeCommand.Additional5GSecurityInformation.GetRINMR()
		}

		ue.UeSecurity.NgKsi.Ksi = int32(message.SecurityModeCommand.SpareHalfOctetAndNgksi.GetNasKeySetIdentifiler())

		// NgKsi: TS 24.501 9.11.3.32
		switch message.SecurityModeCommand.SpareHalfOctetAndNgksi.GetTSC() {
		case nas.TypeOfSecurityContextFlagNative:
			ue.UeSecurity.NgKsi.Tsc = models.ScType_NATIVE
		case nas.TypeOfSecurityContextFlagMapped:
			ue.UeSecurity.NgKsi.Tsc = models.ScType_MAPPED
		}

		// getting NAS Security Mode Complete.
		securityModeComplete, err := mm_5gs.SecurityModeComplete(ue, rinmr)
		if err != nil {
			log.Fatal("[UE][NAS] Error sending Security Mode Complete: ", err)
		}

		// sending to GNB
		sender.SendToGnb(ue, securityModeComplete)
	*/
}

func (ue *UEContext) handleRegistrationAccept(message *nas.RegistrationAccept) {
	/*
		// check the mandatory fields
		if reflect.ValueOf(message.RegistrationAccept.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in Registration Accept, Extended Protocol is missing")
		}

		if message.RegistrationAccept.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in Registration Accept, Extended Protocol not the expected value")
		}

		if message.RegistrationAccept.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Registration Accept, Spare Half not the expected value")
		}

		if message.RegistrationAccept.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in Registration Accept, Security Header not the expected value")
		}

		if reflect.ValueOf(message.RegistrationAccept.RegistrationAcceptMessageIdentity).IsZero() {
			log.Fatal("[UE][NAS] Error in Registration Accept, Message Type is missing")
		}

		if message.RegistrationAccept.RegistrationAcceptMessageIdentity.GetMessageType() != 66 {
			log.Fatal("[UE][NAS] Error in Registration Accept, Message Type not the expected value")
		}

		if reflect.ValueOf(message.RegistrationAccept.RegistrationResult5GS).IsZero() {
			log.Fatal("[UE][NAS] Error in Registration Accept, Registration Result 5GS is missing")
		}

		if message.RegistrationAccept.RegistrationResult5GS.GetRegistrationResultValue5GS() != 1 {
			log.Fatal("[UE][NAS] Error in Registration Accept, Registration Result 5GS not the expected value")
		}

		// change the state of ue for registered
		ue.SetStateMM_REGISTERED()

		// saved 5g GUTI and others information.
		if message.RegistrationAccept.GUTI5G != nil {
			ue.Set5gGuti(message.RegistrationAccept.GUTI5G)
		} else {
			log.Warn("[UE][NAS] UE was not assigned a 5G-GUTI by AMF")
		}

		// use the slice allowed by the network
		// in PDU session request
		if ue.Snssai.Sst == 0 {

			// check the allowed NSSAI received from the 5GC
			snssai := message.RegistrationAccept.AllowedNSSAI.GetSNSSAIValue()

			// update UE slice selected for PDU Session
			ue.Snssai.Sst = int32(snssai[1])
			ue.Snssai.Sd = fmt.Sprintf("0%x0%x0%x", snssai[2], snssai[3], snssai[4])

			log.Warn("[UE][NAS] ALLOWED NSSAI: SST: ", ue.Snssai.Sst, " SD: ", ue.Snssai.Sd)
		}

		log.Info("[UE][NAS] UE 5G GUTI: ", ue.Get5gGuti())

		// getting NAS registration complete.
		registrationComplete, err := mm_5gs.RegistrationComplete(ue)
		if err != nil {
			log.Fatal("[UE][NAS] Error sending Registration Complete: ", err)
		}

		// sending to GNB
		sender.SendToGnb(ue, registrationComplete)
	*/
}

func (ue *UEContext) handleServiceAccept(message *nas.ServiceAccept) {
	// change the state of ue for registered
	ue.SetStateMM_REGISTERED()
}

func (ue *UEContext) handleDlNasTransport(message *nas.DlNasTransport) {

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

func (ue *UEContext) handleIdentityRequest(message *nas.IdentityRequest) {

	switch uint8(message.IdentityType) {
	case nas.MobileIdentity5GSTypeSuci:
		log.Info("[UE][NAS] Requested SUCI 5GS type")
	default:
		log.Fatal("[UE][NAS] Only SUCI identity is supported for now inside MSsim")
	}

	ue.triggerInitIdentifyResponse()
}

func (ue *UEContext) handleConfigurationUpdateCommand(message *nas.ConfigurationUpdateCommand) {
	ue.triggerInitConfigurationUpdateComplete()
}

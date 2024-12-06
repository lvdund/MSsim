package context

import (
	"github.com/reogac/nas"
	log "github.com/sirupsen/logrus"
	context2 "mssim/internal/control_test_engine/gnb/context"
)

const (
	NasCountWindow uint8 = 40
)

func (ue *UEContext) SendGnb(message context2.UEMessage) {
	ue.Lock()
	gnbRx := ue.GetGnbRx()
	if gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		gnbRx <- message
	}
	ue.Unlock()
}

func (ue *UEContext) SendNas(nasPdu []byte) {
	ue.SendGnb(context2.UEMessage{IsNas: true, Nas: nasPdu})
}

func (ue *UEContext) HandleNas(nasPdu []byte) {
	// check if message is null.
	if len(nasPdu) == 0 {
		log.Errorf("[UE][NAS] NAS message is empty")
		return
	}

	var nasMsg nas.NasMessage
	var err error
	if nasMsg, err = nas.Decode(ue.secCtx.NasContext(), nasPdu); err != nil {
		log.Errorf("Decode NasUl failed: %s", err.Error())
		return
	}
	gmm := nasMsg.Gmm
	if gmm == nil {
		log.Errorf("[UE][NAS] NAS message is has no N1MM content")
		return
	}

	switch gmm.MsgType {

	case nas.AuthenticationRequestMsgType:
		// handler authentication request.
		log.Info("[UE][NAS] Receive Authentication Request")
		ue.handlerAuthenticationRequest(gmm.AuthenticationRequest)

	case nas.AuthenticationRejectMsgType:
		// handler authentication reject.
		log.Info("[UE][NAS] Receive Authentication Reject")
		ue.handlerAuthenticationReject(gmm.AuthenticationReject)

	case nas.IdentityRequestMsgType:
		log.Info("[UE][NAS] Receive Identify Request")
		// handler identity request.
		ue.handlerIdentityRequest(gmm.IdentityRequest)

	case nas.SecurityModeCommandMsgType:
		// handler security mode command.
		log.Info("[UE][NAS] Receive Security Mode Command")
		/*
			if !newSecurityContext {
				log.Warn("Received Security Mode Command with security header different from \"Integrity protected with new 5G NAS security context\" ")
			}
		*/
		ue.handlerSecurityModeCommand(gmm.SecurityModeCommand)

	case nas.RegistrationAcceptMsgType:
		// handler registration accept.
		log.Info("[UE][NAS] Receive Registration Accept")
		ue.handlerRegistrationAccept(gmm.RegistrationAccept)

	case nas.ConfigurationUpdateCommandMsgType:
		log.Info("[UE][NAS] Receive Configuration Update Command")
		ue.handlerConfigurationUpdateCommand(gmm.ConfigurationUpdateCommand)

	case nas.DlNasTransportMsgType:
		// handler DL NAS Transport.
		log.Info("[UE][NAS] Receive DL NAS Transport")
		handleCause5GMM(gmm.DlNasTransport.GmmCause)
		ue.handlerDlNasTransport(gmm.DlNasTransport)

	case nas.ServiceAcceptMsgType:
		// handler service reject
		log.Info("[UE][NAS] Receive Service Accept")
		ue.handlerServiceAccept(gmm.ServiceAccept)

	case nas.ServiceRejectMsgType:
		// handler service reject
		log.Error("[UE][NAS] Receive Service Reject")
		handleCause5GMM(&gmm.ServiceReject.GmmCause)

	case nas.RegistrationRejectMsgType:
		// handler registration reject
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
		log.Error("[UE][NAS] UE received a 5GSM Failure, cause: ", cause5GMMToString(uint8(*cause)))
	}
}

func handleCause5GMM(cause *nas.Uint8) {
	if cause != nil {
		log.Error("[UE][NAS] UE received a 5GMM Failure, cause: ", cause5GMMToString(uint8(*cause)))
	}
}

func cause5GMMToString(cause5GMM uint8) string {
	switch cause5GMM {
	case nas.Cause5GMMIllegalUE:
		return "Illegal UE"
	case nas.Cause5GMMPEINotAccepted:
		return "PEI not accepted"
	case nas.Cause5GMMIllegalME:
		return "5GS services not allowed"
	case nas.Cause5GMM5GSServicesNotAllowed:
		return "5GS services not allowed"
	case nas.Cause5GMMUEIdentityCannotBeDerivedByTheNetwork:
		return "UE identity cannot be derived by the network"
	case nas.Cause5GMMImplicitlyDeregistered:
		return "Implicitly de-registered"
	case nas.Cause5GMMPLMNNotAllowed:
		return "PLMN not allowed"
	case nas.Cause5GMMTrackingAreaNotAllowed:
		return "Tracking area not allowed"
	case nas.Cause5GMMRoamingNotAllowedInThisTrackingArea:
		return "Roaming not allowed in this tracking area"
	case nas.Cause5GMMNoSuitableCellsInTrackingArea:
		return "No suitable cells in tracking area"
	case nas.Cause5GMMMACFailure:
		return "MAC failure"
	case nas.Cause5GMMSynchFailure:
		return "Synch failure"
	case nas.Cause5GMMCongestion:
		return "Congestion"
	case nas.Cause5GMMUESecurityCapabilitiesMismatch:
		return "UE security capabilities mismatch"
	case nas.Cause5GMMSecurityModeRejectedUnspecified:
		return "Security mode rejected, unspecified"
	case nas.Cause5GMMNon5GAuthenticationUnacceptable:
		return "Non-5G authentication unacceptable"
	case nas.Cause5GMMN1ModeNotAllowed:
		return "N1 mode not allowed"
	case nas.Cause5GMMRestrictedServiceArea:
		return "Restricted service area"
	case nas.Cause5GMMLADNNotAvailable:
		return "LADN not available"
	case nas.Cause5GMMMaximumNumberOfPDUSessionsReached:
		return "Maximum number of PDU sessions reached"
	case nas.Cause5GMMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nas.Cause5GMMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nas.Cause5GMMngKSIAlreadyInUse:
		return "ngKSI already in use"
	case nas.Cause5GMMNon3GPPAccessTo5GCNNotAllowed:
		return "Non-3GPP access to 5GCN not allowed"
	case nas.Cause5GMMServingNetworkNotAuthorized:
		return "Serving network not authorized"
	case nas.Cause5GMMPayloadWasNotForwarded:
		return "Payload was not forwarded"
	case nas.Cause5GMMDNNNotSupportedOrNotSubscribedInTheSlice:
		return "DNN not supported or not subscribed in the slice"
	case nas.Cause5GMMInsufficientUserPlaneResourcesForThePDUSession:
		return "Insufficient user-plane resources for the PDU session"
	case nas.Cause5GMMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nas.Cause5GMMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nas.Cause5GMMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implementedE"
	case nas.Cause5GMMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nas.Cause5GMMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nas.Cause5GMMConditionalIEError:
		return "Conditional IE error"
	case nas.Cause5GMMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nas.Cause5GMMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please share the pcap with mssim@hpe.com."
	default:
		return "Protocol error, unspecified. Please share the pcap with mssim@hpe.com."
	}
}

func (ue *UEContext) handlerAuthenticationReject(message *nas.AuthenticationReject) {
	/*
		log.Info("[UE][NAS] Authentication of UE ", ue.GetUeId(), " failed")

		ue.SetStateMM_DEREGISTERED()
	*/
}

func (ue *UEContext) handlerAuthenticationRequest(message *nas.AuthenticationRequest) {
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

func (ue *UEContext) handlerSecurityModeCommand(message *nas.SecurityModeCommand) { // check the mandatory fields
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

func (ue *UEContext) handlerRegistrationAccept(message *nas.RegistrationAccept) {
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

func (ue *UEContext) handlerServiceAccept(message *nas.ServiceAccept) {
	// change the state of ue for registered
	//	ue.SetStateMM_REGISTERED()
}

func (ue *UEContext) handlerDlNasTransport(message *nas.DlNasTransport) {
	/*
		// check the mandatory fields
		if reflect.ValueOf(message.DLNASTransport.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Extended Protocol is missing")
		}

		if message.DLNASTransport.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Extended Protocol not expected value")
		}

		if message.DLNASTransport.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Spare Half not expected value")
		}

		if message.DLNASTransport.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Security Header not expected value")
		}

		if message.DLNASTransport.DLNASTRANSPORTMessageIdentity.GetMessageType() != 104 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Message Type is missing or not expected value")
		}

		if reflect.ValueOf(message.DLNASTransport.SpareHalfOctetAndPayloadContainerType).IsZero() {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Payload Container Type is missing")
		}

		if message.DLNASTransport.SpareHalfOctetAndPayloadContainerType.GetPayloadContainerType() != 1 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Payload Container Type not expected value")
		}

		if reflect.ValueOf(message.DLNASTransport.PayloadContainer).IsZero() || message.DLNASTransport.PayloadContainer.GetPayloadContainerContents() == nil {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, Payload Container is missing")
		}

		if reflect.ValueOf(message.DLNASTransport.PduSessionID2Value).IsZero() {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, PDU Session ID is missing")
		}

		if message.DLNASTransport.PduSessionID2Value.GetIei() != 18 {
			log.Fatal("[UE][NAS] Error in DL NAS Transport, PDU Session ID not expected value")
		}

		//getting PDU Session establishment accept.
		payloadContainer := nas_control.GetNasPduFromPduAccept(message)

		switch payloadContainer.GsmHeader.GetMessageType() {
		case nas.MsgTypePDUSessionEstablishmentAccept:
			log.Info("[UE][NAS] Receiving PDU Session Establishment Accept")

			// get UE ip
			pduSessionEstablishmentAccept := payloadContainer.PDUSessionEstablishmentAccept

			// check the mandatory fields
			if reflect.ValueOf(pduSessionEstablishmentAccept.ExtendedProtocolDiscriminator).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, Extended Protocol Discriminator is missing")
			}

			if pduSessionEstablishmentAccept.GetExtendedProtocolDiscriminator() != 46 {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, Extended Protocol Discriminator not expected value")
			}

			if reflect.ValueOf(pduSessionEstablishmentAccept.PDUSessionID).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PDU Session ID is missing or not expected value")
			}

			if reflect.ValueOf(pduSessionEstablishmentAccept.PTI).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PTI is missing")
			}

			if pduSessionEstablishmentAccept.PTI.GetPTI() != 1 {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PTI not the expected value")
			}

			if pduSessionEstablishmentAccept.PDUSESSIONESTABLISHMENTACCEPTMessageIdentity.GetMessageType() != 194 {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, Message Type is missing or not expected value")
			}

			if reflect.ValueOf(pduSessionEstablishmentAccept.SelectedSSCModeAndSelectedPDUSessionType).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, SSC Mode or PDU Session Type is missing")
			}

			if pduSessionEstablishmentAccept.SelectedSSCModeAndSelectedPDUSessionType.GetPDUSessionType() != 1 {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, PDU Session Type not the expected value")
			}

			if reflect.ValueOf(pduSessionEstablishmentAccept.AuthorizedQosRules).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, Authorized QoS Rules is missing")
			}

			if reflect.ValueOf(pduSessionEstablishmentAccept.SessionAMBR).IsZero() {
				log.Fatal("[UE][NAS] Error in PDU Session Establishment Accept, Session AMBR is missing")
			}

			// update PDU Session information.
			pduSessionId := pduSessionEstablishmentAccept.GetPDUSessionID()
			pduSession, err := ue.GetPduSession(pduSessionId)
			if err != nil {
				log.Error("[UE][NAS] Receiving PDU Session Establishment Accept about an unknown PDU Session, id: ", pduSessionId)
				return
			}

			// change the state of ue(SM)(PDU Session Active).
			pduSession.SetStateSM_PDU_SESSION_ACTIVE()

			// get UE IP
			UeIp := pduSessionEstablishmentAccept.GetPDUAddressInformation()
			pduSession.SetIp(UeIp)

			// get QoS Rules
			QosRule := pduSessionEstablishmentAccept.AuthorizedQosRules.GetQosRule()
			// get DNN
			dnn := pduSessionEstablishmentAccept.DNN.GetDNN()
			// get SNSSAI
			sst := pduSessionEstablishmentAccept.SNSSAI.GetSST()
			sd := pduSessionEstablishmentAccept.SNSSAI.GetSD()

			log.Info("[UE][NAS] PDU session QoS RULES: ", QosRule)
			log.Info("[UE][NAS] PDU session DNN: ", string(dnn))
			log.Info("[UE][NAS] PDU session NSSAI -- sst: ", sst, " sd: ",
				fmt.Sprintf("%x%x%x", sd[0], sd[1], sd[2]))
			log.Info("[UE][NAS] PDU address received: ", pduSession.GetIp())
		case nas.MsgTypePDUSessionReleaseCommand:
			log.Info("[UE][NAS] Receiving PDU Session Release Command")

			pduSessionReleaseCommand := payloadContainer.PDUSessionReleaseCommand
			pduSessionId := pduSessionReleaseCommand.GetPDUSessionID()
			pduSession, err := ue.GetPduSession(pduSessionId)
			if pduSession == nil || err != nil {
				log.Error("[UE][NAS] Unable to delete PDU Session ", pduSessionId, " from UE ", ue.GetMsin(), " as the PDU Session was not found. Ignoring.")
				break
			}
			ue.DeletePduSession(pduSessionId)
			log.Info("[UE][NAS] Successfully released PDU Session ", pduSessionId, " from UE Context")
			trigger.InitPduSessionReleaseComplete(ue, pduSession)

		case nas.MsgTypePDUSessionEstablishmentReject:
			log.Error("[UE][NAS] Receiving PDU Session Establishment Reject")

			pduSessionEstablishmentReject := payloadContainer.PDUSessionEstablishmentReject
			pduSessionId := pduSessionEstablishmentReject.GetPDUSessionID()

			log.Error("[UE][NAS] PDU Session Establishment Reject for PDU Session ID ", pduSessionId, ", 5GSM Cause: ", cause5GSMToString(pduSessionEstablishmentReject.GetCauseValue()))

			// Per 5GSM state machine in TS 24.501 - 6.1.3.2.1., we re-try the setup until it's successful
			pduSession, err := ue.GetPduSession(pduSessionId)
			if err != nil {
				log.Error("[UE][NAS] Cannot retry PDU Session Request for PDU Session ", pduSessionId, " after Reject as ", err)
				break
			}
			if pduSession.T3580Retries < 5 {
				// T3580 Timer
				go func() {
					// Exponential backoff
					time.Sleep(time.Duration(math.Pow(5, float64(pduSession.T3580Retries))) * time.Second)
					trigger.InitPduSessionRequestInner(ue, pduSession)
					pduSession.T3580Retries++
				}()
			} else {
				log.Error("[UE][NAS] We re-tried five times to create PDU Session ", pduSessionId, ", Aborting.")
			}

		default:
			log.Error("[UE][NAS] Receiving Unknown Dl NAS Transport message!! ", payloadContainer.GsmHeader.GetMessageType())
		}
	*/
}

func (ue *UEContext) handlerIdentityRequest(message *nas.IdentityRequest) {
	/*
		// check the mandatory fields
		if reflect.ValueOf(message.IdentityRequest.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in Identity Request, Extended Protocol is missing")
		}

		if message.IdentityRequest.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in Identity Request, Extended Protocol not the expected value")
		}

		if message.IdentityRequest.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Identity Request, Spare Half Octet not the expected value")
		}

		if message.IdentityRequest.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in Identity Request, Security Header Type not the expected value")
		}

		if reflect.ValueOf(message.IdentityRequest.IdentityRequestMessageIdentity).IsZero() {
			log.Fatal("[UE][NAS] Error in Identity Request, Message Type is missing")
		}

		if message.IdentityRequest.IdentityRequestMessageIdentity.GetMessageType() != 91 {
			log.Fatal("[UE][NAS] Error in Identity Request, Message Type not the expected value")
		}

		if reflect.ValueOf(message.IdentityRequest.SpareHalfOctetAndIdentityType).IsZero() {
			log.Fatal("[UE][NAS] Error in Identity Request, Spare Half Octet And Identity Type is missing")
		}

		switch message.IdentityRequest.GetTypeOfIdentity() {
		case 1:
			log.Info("[UE][NAS] Requested SUCI 5GS type")
		default:
			log.Fatal("[UE][NAS] Only SUCI identity is supported for now inside MSsim")
		}

		trigger.InitIdentifyResponse(ue)
	*/
}

func (ue *UEContext) handlerConfigurationUpdateCommand(message *nas.ConfigurationUpdateCommand) {
	/*
		// check the mandatory fields
		if reflect.ValueOf(message.ConfigurationUpdateCommand.ExtendedProtocolDiscriminator).IsZero() {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Extended Protocol Discriminator is missing")
		}

		if message.ConfigurationUpdateCommand.ExtendedProtocolDiscriminator.GetExtendedProtocolDiscriminator() != 126 {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Extended Protocol Discriminator not the expected value")
		}

		if message.ConfigurationUpdateCommand.SpareHalfOctetAndSecurityHeaderType.GetSpareHalfOctet() != 0 {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Spare Half not the expected value")
		}

		if message.ConfigurationUpdateCommand.SpareHalfOctetAndSecurityHeaderType.GetSecurityHeaderType() != 0 {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Security Header not the expected value")
		}

		if reflect.ValueOf(message.ConfigurationUpdateCommand.ConfigurationUpdateCommandMessageIdentity).IsZero() {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Message type not the expected value")
		}

		if message.ConfigurationUpdateCommand.ConfigurationUpdateCommandMessageIdentity.GetMessageType() != 84 {
			log.Fatal("[UE][NAS] Error in Configuration Update Command, Message Type not the expected value")
		}

		// return configuration update complete
		trigger.InitConfigurationUpdateComplete(ue)
	*/
}

func cause5GSMToString(causeValue uint8) string {
	switch causeValue {
	case nas.Cause5GSMInsufficientResources:
		return "Insufficient Ressources"
	case nas.Cause5GSMMissingOrUnknownDNN:
		return "Missing or Unknown DNN"
	case nas.Cause5GSMUnknownPDUSessionType:
		return "Unknown PDU Session Type"
	case nas.Cause5GSMUserAuthenticationOrAuthorizationFailed:
		return "User authentification or authorization failed"
	case nas.Cause5GSMRequestRejectedUnspecified:
		return "Request rejected, unspecified"
	case nas.Cause5GSMServiceOptionTemporarilyOutOfOrder:
		return "Service option temporarily out of order."
	case nas.Cause5GSMPTIAlreadyInUse:
		return "PTI already in use"
	case nas.Cause5GSMRegularDeactivation:
		return "Regular deactivation"
	case nas.Cause5GSMReactivationRequested:
		return "Reactivation requested"
	case nas.Cause5GSMInvalidPDUSessionIdentity:
		return "Invalid PDU session identity"
	case nas.Cause5GSMSemanticErrorsInPacketFilter:
		return "Semantic errors in packet filter(s)"
	case nas.Cause5GSMSyntacticalErrorInPacketFilter:
		return "Syntactical error in packet filter(s)"
	case nas.Cause5GSMOutOfLADNServiceArea:
		return "Out of LADN service area"
	case nas.Cause5GSMPTIMismatch:
		return "PTI mismatch"
	case nas.Cause5GSMPDUSessionTypeIPv4OnlyAllowed:
		return "PDU session type IPv4 only allowed"
	case nas.Cause5GSMPDUSessionTypeIPv6OnlyAllowed:
		return "PDU session type IPv6 only allowed"
	case nas.Cause5GSMPDUSessionDoesNotExist:
		return "PDU session does not exist"
	case nas.Cause5GSMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nas.Cause5GSMNotSupportedSSCMode:
		return "Not supported SSC mode"
	case nas.Cause5GSMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nas.Cause5GSMMissingOrUnknownDNNInASlice:
		return "Missing or unknown DNN in a slice"
	case nas.Cause5GSMInvalidPTIValue:
		return "Invalid PTI value"
	case nas.Cause5GSMMaximumDataRatePerUEForUserPlaneIntegrityProtectionIsTooLow:
		return "Maximum data rate per UE for user-plane integrity protection is too low"
	case nas.Cause5GSMSemanticErrorInTheQoSOperation:
		return "Semantic error in the QoS operation"
	case nas.Cause5GSMSyntacticalErrorInTheQoSOperation:
		return "Syntactical error in the QoS operation"
	case nas.Cause5GSMInvalidMappedEPSBearerIdentity:
		return "Invalid mapped EPS bearer identity"
	case nas.Cause5GSMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nas.Cause5GSMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nas.Cause5GSMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implemented"
	case nas.Cause5GSMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nas.Cause5GSMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nas.Cause5GSMConditionalIEError:
		return "Conditional IE error"
	case nas.Cause5GSMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nas.Cause5GSMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please open an issue on Github with pcap."
	default:
		return "Service option temporarily out of order."
	}
}

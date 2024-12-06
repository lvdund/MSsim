package context

import (
	"github.com/reogac/nas"
)

func (ue *UEContext) handlePduSessionEstablishmentAccept(msg *nas.PduSessionEstablishmentAccept) {
	/*
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
	*/
}
func (ue *UEContext) handlePduSessionEstablishmentReject(msg *nas.PduSessionEstablishmentReject) {
	/*
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
	*/
}

func (ue *UEContext) handlePduSessionReleaseCommand(msg *nas.PduSessionReleaseCommand) {
	/*
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
	*/

}

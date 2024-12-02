package ngap

import (
	"bytes"

	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/ies"

	log "github.com/sirupsen/logrus"
)

func Dispatch(amf *context.GNBAmf, gnb *context.GNBContext, message []byte) {

	if message == nil {
		// TODO return error
		log.Info("[GNB][NGAP] NGAP message is nil")
	}

	// decode NGAP message.
	ngapMsg, err, _ := ngap.NgapDecode(bytes.NewReader(message))
	if err != nil {
		log.Error("[GNB][NGAP] Error decoding NGAP message in ", gnb.GetGnbId(), " GNB", ": ", err)
	}

	// check RanUeId and get UE.

	// handle NGAP message.
	switch ngapMsg.Present {

	case ies.NgapPduInitiatingMessage:

		switch ngapMsg.Message.ProcedureCode.Value {

		case ies.ProcedureCode_DownlinkNASTransport:
			// handler NGAP Downlink NAS Transport.
			log.Info("[GNB][NGAP] Receive Downlink NAS Transport")
			HandlerDownlinkNasTransport(gnb, &ngapMsg)

		case ies.ProcedureCode_InitialContextSetup:
			// handler NGAP Initial Context Setup Request.
			log.Info("[GNB][NGAP] Receive Initial Context Setup Request")
			HandlerInitialContextSetupRequest(gnb, &ngapMsg)

		case ies.ProcedureCode_PDUSessionResourceSetup:
			// handler NGAP PDU Session Resource Setup Request.
			log.Info("[GNB][NGAP] Receive PDU Session Resource Setup Request")
			HandlerPduSessionResourceSetupRequest(gnb, &ngapMsg)

		case ies.ProcedureCode_PDUSessionResourceRelease:
			// handler NGAP PDU Session Resource Release
			log.Info("[GNB][NGAP] Receive PDU Session Release Command")
			HandlerPduSessionReleaseCommand(gnb, &ngapMsg)

		case ies.ProcedureCode_UEContextRelease:
			// handler NGAP UE Context Release
			log.Info("[GNB][NGAP] Receive UE Context Release Command")
			HandlerUeContextReleaseCommand(gnb, &ngapMsg)

		case ies.ProcedureCode_AMFConfigurationUpdate:
			// handler NGAP AMF Configuration Update
			log.Info("[GNB][NGAP] Receive AMF Configuration Update")
			HandlerAmfConfigurationUpdate(amf, gnb, &ngapMsg)
		case ies.ProcedureCode_AMFStatusIndication:
			log.Info("[GNB][NGAP] Receive AMF Status Indication")
			HandlerAmfStatusIndication(amf, gnb, &ngapMsg)
		case ies.ProcedureCode_HandoverResourceAllocation:
			// handler NGAP Handover Request
			log.Info("[GNB][NGAP] Receive Handover Request")
			HandlerHandoverRequest(amf, gnb, &ngapMsg)

		case ies.ProcedureCode_Paging:
			// handler NGAP Paging
			log.Info("[GNB][NGAP] Receive Paging")
			HandlerPaging(gnb, &ngapMsg)

		case ies.ProcedureCode_ErrorIndication:
			// handler Error Indicator
			log.Error("[GNB][NGAP] Receive Error Indication")
			HandlerErrorIndication(gnb, &ngapMsg)

		default:
			log.Warnf("[GNB][NGAP] Received unknown NGAP message 0x%x", ngapMsg.Message.ProcedureCode.Value)
		}

	case ies.NgapPduSuccessfulOutcome:

		switch ngapMsg.Message.ProcedureCode.Value {

		case ies.ProcedureCode_NGSetup:
			// handler NGAP Setup Response.
			log.Info("[GNB][NGAP] Receive NG Setup Response")
			HandlerNgSetupResponse(amf, gnb, &ngapMsg)

		case ies.ProcedureCode_PathSwitchRequest:
			// handler PathSwitchRequestAcknowledge
			log.Info("[GNB][NGAP] Receive PathSwitchRequestAcknowledge")
			HandlerPathSwitchRequestAcknowledge(gnb, &ngapMsg)

		case ies.ProcedureCode_HandoverPreparation:
			// handler NGAP AMF Handover Command
			log.Info("[GNB][NGAP] Receive Handover Command")
			HandlerHandoverCommand(amf, gnb, &ngapMsg)

		default:
			log.Warnf("[GNB][NGAP] Received unknown NGAP message 0x%x", ngapMsg.Message.ProcedureCode.Value)
		}

	case ies.NgapPduUnsuccessfulOutcome:

		switch ngapMsg.Message.ProcedureCode.Value {

		case ies.ProcedureCode_NGSetup:
			// handler NGAP Setup Failure.
			log.Info("[GNB][NGAP] Receive Ng Setup Failure")
			HandlerNgSetupFailure(amf, gnb, &ngapMsg)

		default:
			log.Warnf("[GNB][NGAP] Received unknown NGAP message 0x%x", ngapMsg.Message.ProcedureCode.Value)
		}
	}
}

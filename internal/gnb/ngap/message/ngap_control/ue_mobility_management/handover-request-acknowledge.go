package ue_mobility_management

import (
	"bytes"
	"encoding/binary"

	"mssim/internal/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
	"github.com/lvdund/ngap/utils"
	log "github.com/sirupsen/logrus"
)

func HandoverRequestAcknowledge(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	pduSessions := ue.GetPduSessions()
	targetToSourceTransparentContainer := GetTargetToSourceTransparentTransfer()

	msg := &ies.HandoverRequestAcknowledge{
		AMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID: &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		PDUSessionResourceAdmittedList: &ies.PDUSessionResourceAdmittedList{
			Value: make([]*ies.PDUSessionResourceAdmittedItem, len(pduSessions)),
		},
		TargetToSourceTransparentContainer: &ies.TargetToSourceTransparentContainer{
			Value: targetToSourceTransparentContainer,
		},
	}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		//PDU SessionResource Admittedy Item
		PDUSessionID := pduSession.GetPduSessionId()
		HandoverRequestAcknowledgeTransfer := GetHandoverRequestAcknowledgeTransfer(gnb, pduSession)
		msg.PDUSessionResourceAdmittedList.Value = append(msg.PDUSessionResourceAdmittedList.Value, &ies.PDUSessionResourceAdmittedItem{
			PDUSessionID:                       &ies.PDUSessionID{Value: aper.Integer(PDUSessionID)},
			HandoverRequestAcknowledgeTransfer: (*aper.OctetString)(&HandoverRequestAcknowledgeTransfer),
		})
	}

	if len(msg.PDUSessionResourceAdmittedList.Value) == 0 {
		log.Info("[GNB][NGAP] No admitted PDU Session")
	}

	return ngap.NgapEncode(msg)
}

func GetHandoverRequestAcknowledgeTransfer(gnb *context.GNBContext, pduSession *context.GnbPDUSession) []byte {
	data := ies.HandoverRequestAcknowledgeTransfer{}

	downlinkTeid := make([]byte, 4)
	binary.BigEndian.PutUint32(downlinkTeid, pduSession.GetTeidDownlink())
	ip := utils.IPAddressToNgap(gnb.GetN3GnbIp(), "")
	data.DLForwardingUPTNLInformation = &ies.UPTransportLayerInformation{
		Choice: ies.UPTransportLayerInformationPresentGTPTunnel,
		GTPTunnel: &ies.GTPTunnel{
			GTPTEID:               &ies.GTPTEID{Value: downlinkTeid},
			TransportLayerAddress: &ip,
		},
	}

	data.QosFlowFailedToSetupList = &ies.QosFlowListWithCause{
		Value: []*ies.QosFlowWithCauseItem{&ies.QosFlowWithCauseItem{QosFlowIdentifier: &ies.QosFlowIdentifier{Value: 1}}},
	}

	var buf bytes.Buffer
	r := aper.NewWriter(&buf)
	if err := data.Encode(r); err != nil {
		log.Fatalf("aper MarshalWithParams error in GetHandoverRequestAcknowledgeTransfer: %+v", err)
	}
	r.Close()
	return buf.Bytes()
}

func GetTargetToSourceTransparentTransfer() []byte {
	data := ies.TargetNGRANNodeToSourceNGRANNodeTransparentContainer{
		RRCContainer: &ies.RRCContainer{Value: aper.OctetString("\x00\x00\x11")},
	}
	var buf bytes.Buffer
	r := aper.NewWriter(&buf)
	if err := data.Encode(r); err != nil {
		return nil
	}
	r.Close()
	return buf.Bytes()
}

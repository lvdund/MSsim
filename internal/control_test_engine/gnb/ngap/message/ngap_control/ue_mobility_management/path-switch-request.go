package ue_mobility_management

import (
	"bytes"
	"encoding/binary"

	"mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	log "github.com/sirupsen/logrus"

	"github.com/lvdund/ngap/aper"

	"github.com/lvdund/ngap/ies"
	"github.com/lvdund/ngap/utils"
)

func PathSwitchRequest(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	pduSessions := ue.GetPduSessions()

	msg := &ies.PathSwitchRequest{
		SourceAMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID:       &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		PDUSessionResourceToBeSwitchedDLList: &ies.PDUSessionResourceToBeSwitchedDLList{
			Value: []*ies.PDUSessionResourceToBeSwitchedDLItem{},
		},
		UESecurityCapabilities: ue.GetUESecurityCapabilities(),
	}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		ip := utils.IPAddressToNgap(gnb.GetN3GnbIp(), "")
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, pduSession.GetTeidDownlink())
		transfer := ies.PathSwitchRequestTransfer{
			DLNGUUPTNLInformation: &ies.UPTransportLayerInformation{
				Choice: ies.UPTransportLayerInformationPresentGTPTunnel,
				GTPTunnel: &ies.GTPTunnel{
					TransportLayerAddress: &ip,
					GTPTEID:               &ies.GTPTEID{Value: aper.OctetString(buf.Bytes())},
				}},
			DLNGUTNLInformationReused:    nil,
			UserPlaneSecurityInformation: nil,
			QosFlowAcceptedList: &ies.QosFlowAcceptedList{
				Value: []*ies.QosFlowAcceptedItem{{
					QosFlowIdentifier: &ies.QosFlowIdentifier{Value: aper.Integer(pduSession.GetQosId())}}},
			},
		}

		buf = new(bytes.Buffer)
		r := aper.NewWriter(buf)
		if err := transfer.Encode(r); err != nil {
			return nil, err
		}
		r.Close()
		res := aper.OctetString(buf.Bytes())
		msg.PDUSessionResourceToBeSwitchedDLList.Value = append(msg.PDUSessionResourceToBeSwitchedDLList.Value, &ies.PDUSessionResourceToBeSwitchedDLItem{
			PDUSessionID:              &ies.PDUSessionID{Value: aper.Integer(pduSession.GetPduSessionId())},
			PathSwitchRequestTransfer: &res,
		})
	}

	if len(msg.PDUSessionResourceToBeSwitchedDLList.Value) == 0 {
		log.Warnln("[GNB][NGAP] No PDU Session to hand over. Xn Handover requires at least a PDU Session.")
		msg.PDUSessionResourceToBeSwitchedDLList = nil
	}

	return ngap.NgapEncode(msg)
}

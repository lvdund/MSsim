/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2024 Valentin D'Emmanuele
 */
package pdu_session_management

import (
	"bytes"
	"encoding/binary"

	"mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
	"github.com/lvdund/ngap/utils"
)

func PDUSessionResourceSetupResponse(pduSessions []*context.GnbPDUSession, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	msg := &ies.PDUSessionResourceSetupResponse{
		AMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID: &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		PDUSessionResourceSetupListSURes: &ies.PDUSessionResourceSetupListSURes{
			Value: make([]*ies.PDUSessionResourceSetupItemSURes, len(pduSessions)),
		},
	}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		plmnid := pduSession.GetPduSessionId()
		transfer := GetPDUSessionResourceSetupResponseTransfer(gnb.GetN3GnbIp(), pduSession.GetTeidDownlink(), pduSession.GetQosId())
		o := aper.OctetString(transfer)

		msg.PDUSessionResourceFailedToSetupListSURes.Value = append(msg.PDUSessionResourceFailedToSetupListSURes.Value,
			&ies.PDUSessionResourceFailedToSetupItemSURes{
				PDUSessionID: &ies.PDUSessionID{Value: aper.Integer(plmnid)},
				PDUSessionResourceSetupUnsuccessfulTransfer: &o,
			})
	}

	return ngap.NgapEncode(msg)
}

func GetPDUSessionResourceSetupResponseTransfer(ipv4 string, teid uint32, qosId int64) []byte {
	data := ies.PDUSessionResourceSetupResponseTransfer{}

	dowlinkTeid := make([]byte, 4)
	binary.BigEndian.PutUint32(dowlinkTeid, teid)
	ipNgap := utils.IPAddressToNgap(ipv4, "")
	data.DLQosFlowPerTNLInformation = &ies.QosFlowPerTNLInformation{
		UPTransportLayerInformation: &ies.UPTransportLayerInformation{
			Choice: ies.UPTransportLayerInformationPresentGTPTunnel,
			GTPTunnel: &ies.GTPTunnel{
				GTPTEID:               &ies.GTPTEID{Value: dowlinkTeid},
				TransportLayerAddress: &ipNgap,
			},
		},
		AssociatedQosFlowList: &ies.AssociatedQosFlowList{
			Value: []*ies.AssociatedQosFlowItem{
				&ies.AssociatedQosFlowItem{QosFlowIdentifier: &ies.QosFlowIdentifier{Value: aper.Integer(qosId)}},
			},
		},
	}

	var buf bytes.Buffer
	r := aper.NewWriter(&buf)
	if err := data.Encode(r); err != nil {
		return nil
	}
	return buf.Bytes()
}

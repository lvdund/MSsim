/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2024 Valentin D'Emmanuele
 */
package ue_context_management

import (
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
	log "github.com/sirupsen/logrus"
)

func InitialContextSetupResponse(ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	pduSessions := ue.GetPduSessions()
	msg := &ies.InitialContextSetupResponse{
		AMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID: &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		PDUSessionResourceSetupListCxtRes: &ies.PDUSessionResourceSetupListCxtRes{
			Value: []*ies.PDUSessionResourceSetupItemCxtRes{},
		},
	}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		pdusessionid := pduSession.GetPduSessionId()
		transfer := pdu_session_management.GetPDUSessionResourceSetupResponseTransfer(gnb.GetN3GnbIp(), pduSession.GetTeidDownlink(), pduSession.GetQosId())

		msg.PDUSessionResourceSetupListCxtRes.Value = append(msg.PDUSessionResourceSetupListCxtRes.Value,
			&ies.PDUSessionResourceSetupItemCxtRes{
				PDUSessionID:                            &ies.PDUSessionID{Value: aper.Integer(pdusessionid)},
				PDUSessionResourceSetupResponseTransfer: (*aper.OctetString)(&transfer),
			})
	}

	if len(msg.PDUSessionResourceSetupListCxtRes.Value) == 0 {
		log.Info("[GNB][NGAP] No PDU Session to set up in InitialContextSetupResponse.")
		msg.PDUSessionResourceSetupListCxtRes = nil
	}

	return ngap.NgapEncode(msg)
}

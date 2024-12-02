/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Valentin D'Emmanuele
 */
package ue_context_management

import (
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"

	"github.com/lvdund/ngap/ies"
)

func UeContextReleaseRequest(ue *context.GNBUe) ([]byte, error) {
	msg := &ies.UEContextReleaseRequest{
		AMFUENGAPID:                     &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID:                     &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		PDUSessionResourceListCxtRelReq: &ies.PDUSessionResourceListCxtRelReq{},
	}

	activePduSession := []*context.GnbPDUSession{}
	pduSessions := ue.GetPduSessions()
	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		activePduSession = append(activePduSession, pduSession)
	}

	if len(activePduSession) > 0 {
		msg.PDUSessionResourceListCxtRelReq = &ies.PDUSessionResourceListCxtRelReq{
			Value: make([]*ies.PDUSessionResourceItemCxtRelReq, len(activePduSession)),
		}

		// PDU Session Resource Item in PDU session Resource List
		for _, pduSessionID := range activePduSession {
			id := pduSessionID.GetPduSessionId()
			msg.PDUSessionResourceListCxtRelReq.Value = append(msg.PDUSessionResourceListCxtRelReq.Value,
				&ies.PDUSessionResourceItemCxtRelReq{
					PDUSessionID: &ies.PDUSessionID{Value: aper.Integer(id)},
				})
		}
	}
	return ngap.NgapEncode(msg)
}

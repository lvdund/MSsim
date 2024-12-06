package pdu_session_management

import (
	"mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

func PDUSessionReleaseResponse(pduSessionIds []ies.PDUSessionID, ue *context.GNBUe) ([]byte, error) {
	amfUeNgapID := ue.GetAmfId()
	ranUeNgapID := ue.GetRanUeId()
	msg := ies.PDUSessionResourceReleaseResponse{}

	msg.AMFUENGAPID = &ies.AMFUENGAPID{Value: aper.Integer(amfUeNgapID)}

	msg.RANUENGAPID = &ies.RANUENGAPID{Value: aper.Integer(ranUeNgapID)}

	msg.PDUSessionResourceReleasedListRelRes = &ies.PDUSessionResourceReleasedListRelRes{
		Value: make([]*ies.PDUSessionResourceReleasedItemRelRes, len(pduSessionIds)),
	}
	for _, pduSessionId := range pduSessionIds {
		msg.PDUSessionResourceReleasedListRelRes.Value = append(msg.PDUSessionResourceReleasedListRelRes.Value, &ies.PDUSessionResourceReleasedItemRelRes{
			PDUSessionID: &pduSessionId,
			PDUSessionResourceReleaseResponseTransfer: &aper.OctetString{00},
		})
	}
	return ngap.NgapEncode(&msg)
}

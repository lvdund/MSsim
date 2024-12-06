package nas_transport

import (
	"fmt"

	"mssim/internal/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

func getUplinkNASTransport(amfUeNgapID, ranUeNgapID int64, nasPdu []byte, gnb *context.GNBContext) ([]byte, error) {
	msg := ies.UplinkNASTransport{}

	// AMF UE NGAP ID
	msg.AMFUENGAPID = &ies.AMFUENGAPID{Value: aper.Integer(amfUeNgapID)}

	// RAN UE NGAP ID
	msg.RANUENGAPID = &ies.RANUENGAPID{Value: aper.Integer(ranUeNgapID)}

	// NAS-PDU
	msg.NASPDU = &ies.NASPDU{Value: nasPdu}

	// User Location Information
	plmnid := gnb.GetPLMNIdentity()
	cellid := gnb.GetNRCellIdentity()
	tac := gnb.GetTacInBytes()
	msg.UserLocationInformation = &ies.UserLocationInformation{
		Choice: ies.UserLocationInformationPresentUserLocationInformationNR,
		UserLocationInformationNR: &ies.UserLocationInformationNR{
			NRCGI: &ies.NRCGI{
				PLMNIdentity:   &plmnid,
				NRCellIdentity: &cellid,
			},
			TAI: &ies.TAI{
				PLMNIdentity: &plmnid,
				TAC:          &ies.TAC{Value: tac},
			},
		},
	}
	return ngap.NgapEncode(&msg)
}

func SendUplinkNasTransport(message []byte, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {

	sendMsg, err := getUplinkNASTransport(ue.GetAmfUeId(), ue.GetRanUeId(), message, gnb)
	if err != nil {
		return nil, fmt.Errorf("Error getting UE Id %d NAS Authentication Response", ue.GetRanUeId())
	}

	return sendMsg, nil
}

package nas_transport

import (
	"fmt"

	"mssim/internal/gnb/context"

	"github.com/reogac/nas"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

var TestPlmn ies.PLMNIdentity

func init() {
	// TODO PLMN is hardcode here.
	TestPlmn.Value = aper.OctetString("\x02\xf8\x39")
}

func GetInitialUEMessage(ranUeNgapID int64, nasPdu []byte, guti5g *nas.Guti, gnb *context.GNBContext) ([]byte, error) {
	msg := ies.InitialUEMessage{}

	msg.RANUENGAPID = &ies.RANUENGAPID{
		Value: aper.Integer(ranUeNgapID),
	}

	msg.NASPDU = &ies.NASPDU{Value: nasPdu}

	plmnid := gnb.GetPLMNIdentity()
	cellid := gnb.GetNRCellIdentity()
	plmnid_tai := gnb.GetMccAndMncInOctets()
	tac := gnb.GetTacInBytes()
	msg.UserLocationInformation = &ies.UserLocationInformation{
		Choice: ies.UserLocationInformationPresentUserLocationInformationNR,
		UserLocationInformationNR: &ies.UserLocationInformationNR{
			NRCGI: &ies.NRCGI{
				PLMNIdentity:   &plmnid,
				NRCellIdentity: &cellid,
			},
			TAI: &ies.TAI{
				PLMNIdentity: &ies.PLMNIdentity{Value: plmnid_tai},
				TAC:          &ies.TAC{Value: tac},
			},
		},
	}

	msg.RRCEstablishmentCause = &ies.RRCEstablishmentCause{Value: ies.RRCEstablishmentCauseMosignalling}

	// 5G-S-TSMI (optional)
	if guti5g != nil {
		tmsi := guti5g.GetTMSI5G()
		msg.FiveGSTMSI = &ies.FiveGSTMSI{
			AMFSetID: &ies.AMFSetID{Value: aper.BitString{
				Bytes:   []byte{guti5g.Octet[5], guti5g.Octet[6]},
				NumBits: 10,
			}},
			AMFPointer: &ies.AMFPointer{Value: aper.BitString{
				Bytes:   []byte{guti5g.GetAMFPointer()},
				NumBits: 6,
			}},
			FiveGTMSI: &ies.FiveGTMSI{Value: tmsi[:]},
		}
	}

	// UE Context Request (optional)
	msg.UEContextRequest = &ies.UEContextRequest{Value: ies.UEContextRequestRequested}

	return ngap.NgapEncode(&msg)
}

func SendInitialUeMessage(registrationRequest []byte, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	sendMsg, err := GetInitialUEMessage(ue.GetRanUeId(), registrationRequest, ue.GetTMSI(), gnb)
	if err != nil {
		return nil, fmt.Errorf("Error in %d ue initial message", ue.GetRanUeId())
	}

	return sendMsg, nil
}

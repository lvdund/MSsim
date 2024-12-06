package ue_mobility_management

import (
	"bytes"

	"mssim/internal/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"

	log "github.com/sirupsen/logrus"
)

func HandoverRequired(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	pduSessions := ue.GetPduSessions()
	PLMNIdentity := targetGnb.GetPLMNIdentity()
	TAC := targetGnb.GetTacInBytes()
	transfer := GetSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, ue.GetPrUeId())

	msg := &ies.HandoverRequired{
		AMFUENGAPID:  &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID:  &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		HandoverType: &ies.HandoverType{Value: ies.HandoverTypeIntra5Gs},
		Cause: &ies.Cause{
			Choice:       ies.CausePresentRadioNetwork,
			RadioNetwork: &ies.CauseRadioNetwork{Value: ies.CauseRadioNetworkHandoverdesirableforradioreason},
		},
		PDUSessionResourceListHORqd: &ies.PDUSessionResourceListHORqd{
			Value: make([]*ies.PDUSessionResourceItemHORqd, len(pduSessions)),
		},
		TargetID: &ies.TargetID{
			Choice: ies.TargetIDPresentTargetRANNodeID,
			TargetRANNodeID: &ies.TargetRANNodeID{
				GlobalRANNodeID: &ies.GlobalRANNodeID{
					Choice: ies.GlobalRANNodeIDPresentGlobalGNBID,
					GlobalGNBID: &ies.GlobalGNBID{
						PLMNIdentity: &PLMNIdentity,
						GNBID: &ies.GNBID{
							Choice: ies.GNBIDPresentGNBID,
							GNBID: &aper.BitString{
								Bytes:   targetGnb.GetGnbIdInBytes(),
								NumBits: uint64(len(targetGnb.GetGnbIdInBytes()) * 8),
							},
						},
					},
				},
				SelectedTAI: &ies.TAI{
					PLMNIdentity: &PLMNIdentity,
					TAC:          &ies.TAC{Value: TAC},
				},
			},
		},
		SourceToTargetTransparentContainer: &ies.SourceToTargetTransparentContainer{Value: transfer},
	}

	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		//PDU SessionResource Admittedy Item
		PDUSessionID := pduSession.GetPduSessionId()

		transfer := ies.HandoverRequiredTransfer{}
		var buf bytes.Buffer
		r := aper.NewWriter(&buf)
		if err := transfer.Encode(r); err != nil {
			log.Warnln("[GNB][NGAP] err encode HandoverRequiredBuilder <- HandoverRequiredTransfer ")
		}
		res := buf.Bytes()

		msg.PDUSessionResourceListHORqd.Value = append(msg.PDUSessionResourceListHORqd.Value,
			&ies.PDUSessionResourceItemHORqd{
				PDUSessionID:             &ies.PDUSessionID{Value: aper.Integer(PDUSessionID)},
				HandoverRequiredTransfer: (*aper.OctetString)(&res),
			})
	}

	if len(msg.PDUSessionResourceListHORqd.Value) == 0 {
		log.Error("[GNB][NGAP] No PDU Session to set up in InitialContextSetupResponse. NGAP Handover requires at least a PDU Session.")
	}

	return ngap.NgapEncode(msg)
}

func GetSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) []byte {
	data := buildSourceToTargetTransparentTransfer(sourceGnb, targetGnb, pduSessions, prUeId)
	var buf bytes.Buffer
	r := aper.NewWriter(&buf)
	if err := data.Encode(r); err != nil {
		log.Fatalf("aper MarshalWithParams error in GetSourceToTargetTransparentTransfer: %+v", err)
	}
	return buf.Bytes()
}

func buildSourceToTargetTransparentTransfer(sourceGnb *context.GNBContext, targetGnb *context.GNBContext, pduSessions [16]*context.GnbPDUSession, prUeId int64) (data ies.SourceNGRANNodeToTargetNGRANNodeTransparentContainer) {
	data = ies.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}

	data.RRCContainer = &ies.RRCContainer{Value: aper.OctetString("\x00\x00\x11")}
	data.IndexToRFSP = &ies.IndexToRFSP{Value: aper.Integer(prUeId)}

	data.PDUSessionResourceInformationList = new(ies.PDUSessionResourceInformationList)
	for _, pduSession := range pduSessions {
		if pduSession == nil {
			continue
		}
		data.PDUSessionResourceInformationList.Value = append(data.PDUSessionResourceInformationList.Value, &ies.PDUSessionResourceInformationItem{
			PDUSessionID: &ies.PDUSessionID{Value: aper.Integer(pduSession.GetPduSessionId())},
			QosFlowInformationList: &ies.QosFlowInformationList{
				Value: []*ies.QosFlowInformationItem{
					&ies.QosFlowInformationItem{
						QosFlowIdentifier: &ies.QosFlowIdentifier{Value: 1},
					},
				},
			},
		})
	}
	if len(data.PDUSessionResourceInformationList.Value) == 0 {
		log.Error("[GNB][NGAP] No PDU Session to set up in InitialContextSetupResponse. NGAP Handover requires at least a PDU Session.")
		data.PDUSessionResourceInformationList = nil
	}

	PLMNIdentity := targetGnb.GetPLMNIdentity()
	NRCellIdentity := targetGnb.GetNRCellIdentity()
	data.TargetCellID = &ies.NGRANCGI{
		Choice: ies.TargetIDPresentTargetRANNodeID,
		NRCGI: &ies.NRCGI{
			PLMNIdentity:   &PLMNIdentity,
			NRCellIdentity: &NRCellIdentity,
		},
	}

	PLMNIdentity = sourceGnb.GetPLMNIdentity()
	NRCellIdentity = sourceGnb.GetNRCellIdentity()
	data.UEHistoryInformation = &ies.UEHistoryInformation{
		Value: []*ies.LastVisitedCellItem{
			&ies.LastVisitedCellItem{
				LastVisitedCellInformation: &ies.LastVisitedCellInformation{
					Choice: ies.LastVisitedCellInformationPresentNGRANCell,
					NGRANCell: &ies.LastVisitedNGRANCellInformation{
						GlobalCellID: &ies.NGRANCGI{
							Choice: ies.NGRANCGIPresentNRCGI,
							NRCGI: &ies.NRCGI{
								PLMNIdentity:   &PLMNIdentity,
								NRCellIdentity: &NRCellIdentity,
							},
						},
						CellType: &ies.CellType{
							CellSize: &ies.CellSize{Value: ies.CellSizeVerysmall},
						},
						TimeUEStayedInCell: &ies.TimeUEStayedInCell{Value: 10},
					},
				},
			},
		},
	}
	return data
}

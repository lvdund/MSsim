package interface_management

import (
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"

	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

func NGSetupRequest(gnb *context.GNBContext, name string) ([]byte, error) {

	msg := ies.NGSetupRequest{}

	msg.GlobalRANNodeID = &ies.GlobalRANNodeID{
		Choice: ies.GlobalRANNodeIDPresentGlobalGNBID,
		GlobalGNBID: &ies.GlobalGNBID{
			PLMNIdentity: &ies.PLMNIdentity{Value: gnb.GetMccAndMncInOctets()},
			GNBID: &ies.GNBID{
				Choice: ies.GNBIDPresentGNBID,
				GNBID: &aper.BitString{
					Bytes:   gnb.GetGnbIdInBytes(),
					NumBits: 24,
				},
			},
		},
	}

	msg.RANNodeName = &ies.RANNodeName{Value: aper.OctetString(name)}

	sst, sd := gnb.GetSliceInBytes()
	msg.SupportedTAList = &ies.SupportedTAList{
		Value: []*ies.SupportedTAItem{
			&ies.SupportedTAItem{
				TAC: &ies.TAC{Value: gnb.GetTacInBytes()},
				BroadcastPLMNList: &ies.BroadcastPLMNList{
					Value: []*ies.BroadcastPLMNItem{
						&ies.BroadcastPLMNItem{
							PLMNIdentity: &ies.PLMNIdentity{Value: gnb.GetMccAndMncInOctets()},
							TAISliceSupportList: &ies.SliceSupportList{
								Value: []*ies.SliceSupportItem{
									&ies.SliceSupportItem{
										SNSSAI: &ies.SNSSAI{
											SST: &ies.SST{Value: sst},
											SD:  &ies.SD{Value: sd},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	msg.DefaultPagingDRX = &ies.PagingDRX{
		Value: ies.PagingDRXV128,
	}

	return ngap.NgapEncode(&msg)
}

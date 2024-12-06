/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Valentin D'Emmanuele
 */
package ue_mobility_management

import (
	"mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

func HandoverNotify(gnb *context.GNBContext, ue *context.GNBUe) ([]byte, error) {
	PLMNIdentity := gnb.GetPLMNIdentity()
	NRCellIdentity := gnb.GetNRCellIdentity()
	TAC := gnb.GetTacInBytes()

	msg := &ies.HandoverNotify{
		AMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID: &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
		UserLocationInformation: &ies.UserLocationInformation{
			Choice: ies.UserLocationInformationPresentUserLocationInformationNR,
			UserLocationInformationNR: &ies.UserLocationInformationNR{
				NRCGI: &ies.NRCGI{
					PLMNIdentity:   &PLMNIdentity,
					NRCellIdentity: &NRCellIdentity,
				},
				TAI: &ies.TAI{
					PLMNIdentity: &PLMNIdentity,
					TAC:          &ies.TAC{Value: TAC},
				},
			},
		},
	}

	return ngap.NgapEncode(msg)
}

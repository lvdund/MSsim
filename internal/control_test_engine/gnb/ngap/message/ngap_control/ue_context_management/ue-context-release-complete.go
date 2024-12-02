package ue_context_management

import (
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
)

func UeContextReleaseComplete(ue *context.GNBUe) ([]byte, error) {
	msg := ies.UEContextReleaseComplete{
		AMFUENGAPID: &ies.AMFUENGAPID{Value: aper.Integer(ue.GetAmfUeId())},
		RANUENGAPID: &ies.RANUENGAPID{Value: aper.Integer(ue.GetRanUeId())},
	}
	return ngap.NgapEncode(&msg)
}

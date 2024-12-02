package state

import (
	"github.com/lvdund/mssim/internal/control_test_engine/ue/context"
	"github.com/lvdund/mssim/internal/control_test_engine/ue/nas"
)

func DispatchState(ue *context.UEContext, message []byte) {
	nas.DispatchNas(ue, message)
}

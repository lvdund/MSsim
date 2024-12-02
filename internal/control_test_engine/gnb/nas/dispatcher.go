package nas

import (
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/nas/handler"
)

func Dispatch(ue *context.GNBUe, message []byte, gnb *context.GNBContext) {

	switch ue.GetState() {

	case context.Initialized:
		// handler UE message.
		handler.HandlerUeInitialized(ue, message, gnb)

	case context.Ongoing:
		// handler UE message.
		handler.HandlerUeOngoing(ue, message, gnb)

	case context.Ready:
		// handler UE message.
		handler.HandlerUeReady(ue, message, gnb)
	}
}

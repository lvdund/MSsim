package procedures

import "github.com/lvdund/mssim/internal/control_test_engine/gnb/context"

type UeTesterMessageType int32

const (
	Registration      UeTesterMessageType = 0
	Deregistration    UeTesterMessageType = 1
	NewPDUSession     UeTesterMessageType = 2
	DestroyPDUSession UeTesterMessageType = 3
	Terminate         UeTesterMessageType = 4
	Kill              UeTesterMessageType = 5
	Idle              UeTesterMessageType = 6
	ServiceRequest    UeTesterMessageType = 7
)

type UeTesterMessage struct {
	Type    UeTesterMessageType
	Param   uint8
	GnbChan chan context.UEMessage
}
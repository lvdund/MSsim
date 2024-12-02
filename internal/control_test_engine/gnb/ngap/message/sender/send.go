package sender

import (
	"fmt"

	"github.com/lvdund/mssim/lib/ngap/ngapSctp"

	"github.com/ishidawataru/sctp"
)

func SendToAmF(message []byte, conn *sctp.SCTPConn) error {

	// TODO included information for SCTP association.
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}

	_, err := conn.SCTPWrite(message, info)
	if err != nil {
		return fmt.Errorf("Error sending NGAP message ", err)
	}

	return nil
}

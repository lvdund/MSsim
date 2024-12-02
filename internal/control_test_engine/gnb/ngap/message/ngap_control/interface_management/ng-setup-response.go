package interface_management

import (
	"bytes"
	"fmt"

	"github.com/ishidawataru/sctp"
	"github.com/lvdund/ngap"
)

func NgSetupResponse(connN2 *sctp.SCTPConn) (*ngap.NgapPdu, error) {
	var recvMsg = make([]byte, 2048)
	var n int

	// receive NGAP message from AMF.
	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %s NG-SETUP-RESPONSE")
	}

	// ngapMsg, err, _ := ngap.NgapDecode(recvMsg[:n])
	ngapMsg, err, _ := ngap.NgapDecode(bytes.NewBuffer(recvMsg[:n]))
	if err != nil {
		return nil, fmt.Errorf("Error decoding %s NG-SETUP-RESPONSE")
	}

	return &ngapMsg, nil
}

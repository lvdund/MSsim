package pdu_session_management

import (
	"fmt"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/ies"
)

func GetGtpTeid(ngap *ngap.NgapPdu) (gtpTeid []byte, err error) {
	msg := ngap.Message.Msg.(*ies.PDUSessionResourceSetupRequest)
	var tranfer []byte
	if msg.PDUSessionResourceSetupListSUReq != nil {
		for _, ie := range msg.PDUSessionResourceSetupListSUReq.Value {
			tranfer = *ie.PDUSessionResourceSetupRequestTransfer
			break
		}
	}
	ie := ies.PDUSessionResourceSetupRequestTransfer{}
	if err, _ = ie.Decode(tranfer); err != nil {
		return
	}
	if ie.ULNGUUPTNLInformation != nil {
		gtpTeid := ie.ULNGUUPTNLInformation.GTPTunnel.GTPTEID.Value
		return gtpTeid, nil
	}
	return gtpTeid, fmt.Errorf("ULNGUUPTNLInformation is nil")
}

package sm_5gs

import (
	"fmt"

	"github.com/free5gc/nas"
	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/ies"

	"github.com/lvdund/mssim/internal/control_test_engine/ue/nas/message/nas_control"
)

func DecodeNasPduAccept(ngapMsg *ngap.NgapPdu) (*nas.Message, error) {

	// get NasPdu from DlNas.
	nasPdu := nas_control.GetNasPduFromDlNas(ngapMsg.Message.Msg.(*ies.PDUSessionResourceSetupRequest))
	if nasPdu == nil {
		return nil, fmt.Errorf("Error in get NasPdu from DL NAS message")
	}

	// get NasPdu from Pdu Session establishment accept.
	nasPduPayload := nas_control.GetNasPduFromPduAccept(nasPdu)
	if nasPduPayload == nil {
		return nil, fmt.Errorf("Error in get NasPdu from Pdu Session establishment accept message")
	}

	return nasPduPayload, nil
}

func GetPduAdress(m *nas.Message) [12]uint8 {
	return m.PDUSessionEstablishmentAccept.GetPDUAddressInformation()
}

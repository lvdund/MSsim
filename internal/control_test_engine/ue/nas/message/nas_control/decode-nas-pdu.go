package nas_control

import (
	"github.com/free5gc/nas"
	"github.com/lvdund/ngap/ies"
)

/*
func GetNasPduFromDownlink(msg *ies.DownlinkNASTransport) (m *nas.Message) {
	if msg.NASPDU != nil {
		pkg := []byte(msg.NASPDU.Value)
		m = new(nas.Message)
		err := m.PlainNasDecode(&pkg)
		if err != nil {
			return nil
		}
		return
	}
	return nil
}

func GetNasPduFromPduAccept(dlNas *nas.Message) (m *nas.Message) {

	// get payload container from DL NAS.
	payload := dlNas.DLNASTransport.GetPayloadContainerContents()
	m = new(nas.Message)
	err := m.PlainNasDecode(&payload)
	if err != nil {
		return nil
	}
	return
}

func GetNasPduFromDlNas(msg *ies.PDUSessionResourceSetupRequest) (m *nas.Message) {
	if msg.PDUSessionResourceSetupListSUReq != nil {
		pDUSessionResourceSetupList := msg.PDUSessionResourceSetupListSUReq
		for _, item := range pDUSessionResourceSetupList.Value {
			// get PDUSessionNas-PDU
			payload := []byte(item.PDUSessionNASPDU.Value)
			// remove security header.
			payload = payload[7:]
			m := new(nas.Message)
			err := m.PlainNasDecode(&payload)
			if err != nil {
				return nil
			}
			return m
		}
	}
	return nil
}
*/

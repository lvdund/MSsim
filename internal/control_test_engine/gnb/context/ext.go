package context

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"github.com/lvdund/mssim/lib/ngap/ngapSctp"
)

func (ue *GNBUe) ReceiveMessage(msg *UEMessage) {
	ue.Lock()
	gnbTx := ue.GetGnbTx()
	if gnbTx == nil {
		//log.Warn("[GNB] Do not send NAS messages to UE as channel is closed")
	} else {
		gnbTx <- *msg
	}
	ue.Unlock()

}

func (ue *GNBUe) ReceiveNas(nasPdu []byte) {
	ue.ReceiveMessage(&UEMessage{
		IsNas: true,
		Nas:   nasPdu,
	})
}

func (ue *GNBUe) SendNgap(pdu []byte) error {
	// TODO included information for SCTP association.
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}

	_, err := ue.sctpConnection.SCTPWrite(pdu, info)
	if err != nil {
		return fmt.Errorf("Error sending NGAP message ", err)
	}

	return nil

}

func (ue *GNBUe) SendNas(nasPdu []byte, gnb *GNBContext) {
	var ngap []byte
	var err error
	newState := ue.state
	switch ue.GetState() {

	case Initialized:
		//ngap, err := nas_transport.SendInitialUeMessage(message, ue, gnb)
		newState = Ongoing

	case Ongoing:
		//ngap, err = nas_transport.SendUplinkNasTransport(message, ue, gnb)

	case Ready:
		//ngap, err = nas_transport.SendUplinkNasTransport(message, ue, gnb)
	}
	if err != nil {
		//log.Errorln("[GNB][NGAP] Error making Uplink Nas Transport: ", err)
		return
	}
	ue.state = newState
	err = ue.SendNgap(ngap)
	if err != nil {
		//log.Errorln("[GNB][AMF] Error sending Nas message in NGAP: ", err)
	}

}

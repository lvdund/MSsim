package mm_5gs

/*
import (
	"bytes"
	"fmt"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"mssim/internal/control_test_engine/ue/context"
)

func IdentityResponse(ue *context.UEContext) (nasPdu []byte) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeIdentityResponse)

	identityResponse := nasMessage.NewIdentityResponse(0)
	identityResponse.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	identityResponse.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	identityResponse.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0x00)
	identityResponse.IdentityResponseMessageIdentity.SetMessageType(nas.MsgTypeIdentityResponse)
	identityResponse.MobileIdentity = nasType.MobileIdentity(ue.GetSuci())

	m.GmmMessage.IdentityResponse = identityResponse

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
*/

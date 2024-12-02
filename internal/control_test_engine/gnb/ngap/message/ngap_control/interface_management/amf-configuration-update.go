package interface_management

import (
	"github.com/lvdund/ngap"

	"github.com/lvdund/ngap/ies"
)

func AmfConfigurationUpdateAcknowledge() ([]byte, error) {
	message := ies.AMFConfigurationUpdateAcknowledge{}

	return ngap.NgapEncode(&message)
}

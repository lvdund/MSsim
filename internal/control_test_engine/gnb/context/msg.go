package context

import "github.com/free5gc/nas/nasType"

type UEMessage struct {
	GNBPduSessions    [16]*GnbPDUSession
	GnbIp             string
	GNBRx             chan UEMessage
	GNBTx             chan UEMessage
	GNBInboundChannel chan UEMessage
	IsNas             bool
	Nas               []byte
	ConnectionClosed  bool
	PrUeId            int64
	Tmsi              *nasType.GUTI5G
	Mcc               string
	Mnc               string
	UEContext         *GNBUe
	IsHandover        bool
	Idle              bool
	FetchPagedUEs     bool
	PagedUEs          []PagedUE
}

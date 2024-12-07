package context

import "github.com/reogac/nas"

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
	Tmsi              *nas.Guti
	Mcc               string
	Mnc               string
	UEContext         *GNBUe
	IsHandover        bool
	Idle              bool
	FetchPagedUEs     bool
	PagedUEs          []PagedUE
}

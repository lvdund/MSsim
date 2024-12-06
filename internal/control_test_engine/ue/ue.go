package ue

import (
	"sync"

	"mssim/config"
	gnbContext "mssim/internal/control_test_engine/gnb/context"
	"mssim/internal/control_test_engine/procedures"
	"mssim/internal/control_test_engine/ue/context"
)

func NewUE(conf config.Config, id int, ueMgrChannel chan procedures.UeTesterMessage, gnbInboundChannel chan gnbContext.UEMessage, wg *sync.WaitGroup, logFile string) chan context.ScenarioMessage {
	// new UE instance.
	ue := &context.UEContext{}
	scenarioChan := make(chan context.ScenarioMessage)

	// new UE context
	ue.NewRanUeContext(
		conf.Ue.Msin,
		conf.Ue.GetUESecurityCapability(),
		conf.Ue.Key,
		conf.Ue.Opc,
		"c9e8763286b5b9ffbdf56e1297d0887b",
		conf.Ue.Amf,
		conf.Ue.Sqn,
		conf.Ue.Hplmn.Mcc,
		conf.Ue.Hplmn.Mnc,
		conf.Ue.RoutingIndicator,
		conf.Ue.Dnn,
		int32(conf.Ue.Snssai.Sst),
		conf.Ue.Snssai.Sd,
		conf.Ue.TunnelMode,
		scenarioChan,
		gnbInboundChannel,
		id, logFile)

	go ue.Service(wg, ueMgrChannel)
	return scenarioChan
}

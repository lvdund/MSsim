package templates

import (
	"sync"

	"mssim/config"
	"mssim/internal/gnb"
)

func TestAttachGnbWithConfiguration() {

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// wrong messages:
	// cfg.GNodeB.PlmnList.Mcc = "891"
	// cfg.GNodeB.PlmnList.Mnc = "23"
	// cfg.GNodeB.PlmnList.Tac = "000002"
	// cfg.GNodeB.SliceSupportList.St = "10"
	// cfg.GNodeB.SliceSupportList.Sst = "010239"

	go gnb.InitGnb(cfg, &wg)

	wg.Add(1)

	wg.Wait()
}

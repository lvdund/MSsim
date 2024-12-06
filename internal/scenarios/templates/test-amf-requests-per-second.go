package templates

import (
	"strconv"
	"sync"

	"mssim/config"
	"mssim/internal/control_test_engine/gnb"
	"mssim/monitoring"

	log "github.com/sirupsen/logrus"
)

// rajada de mensagens por segundo enviadas
// para o AMF
// durante um período de tempo
func TestRqsLoop(numRqs int, interval int) int64 {

	wg := sync.WaitGroup{}

	monitor := monitoring.Monitor{
		RqsL: 0,
		RqsG: 0,
	}

	cfg := config.GetConfig()

	ranPort := 1000
	for y := 1; y <= interval; y++ {

		monitor.InitRqsLocal()

		for i := 1; i <= numRqs; i++ {

			cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)

			cfg.GNodeB.ControlIF.Port = ranPort

			go gnb.InitGnbForLoadSeconds(cfg, &wg, &monitor)

			wg.Add(1)

			ranPort++
		}

		wg.Wait()

		log.Warn("[TESTER][GNB] AMF Responses per Second:", monitor.GetRqsLocal())
		monitor.SetRqsGlobal(monitor.GetRqsLocal())
	}

	return monitor.GetRqsGlobal()
}

func gnbIdGenerator(i int) string {

	var base string
	switch true {
	case i < 10:
		base = "00000"
	case i < 100:
		base = "0000"
	case i >= 100:
		base = "000"
	}

	gnbId := base + strconv.Itoa(i)
	return gnbId
}

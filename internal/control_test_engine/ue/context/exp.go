package context

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Experiment struct {
	CreatedTime   time.Time
	ActivatedTime time.Time
}

func (exp *Experiment) Duration() int64 {
	return int64(exp.ActivatedTime.Sub(exp.CreatedTime))
}

func LogExpResults(logFile *os.File, ue *UEContext) {
	if ue == nil || len(ue.PduSession) == 0 {
		return
	} else if logFile == nil {
		logrus.Errorln("logFile is nil -", ue.UeSecurity.Supi)
	}
	for _, pdu := range ue.PduSession {
		if pdu != nil {
			fmt.Fprintf(logFile, "%s-%d: %d\n", ue.UeSecurity.Supi, pdu.Id, pdu.Exp.Duration())
		}
	}
}

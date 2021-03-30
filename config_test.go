package bitxid

import (
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/sirupsen/logrus"
)

const (
	loggerAccountDID = "account-did"
	loggerChainDID   = "chain-did"
)

var w *loggerWrapper

type loggerWrapper struct {
	loggers map[string]*logrus.Entry
}

// Init initializes all loggers
func loggerInit() {
	// config *repo.Config
	m := make(map[string]*logrus.Entry)
	m[loggerAccountDID] = log.NewWithModule(loggerAccountDID)
	m[loggerAccountDID].Logger.SetLevel(log.ParseLevel("info"))
	m[loggerChainDID] = log.NewWithModule(loggerChainDID)
	m[loggerChainDID].Logger.SetLevel(log.ParseLevel("info"))

	w = &loggerWrapper{loggers: m}
}

// Get gets logger for specific module
func loggerGet(name string) logrus.FieldLogger {
	return w.loggers[name]
}

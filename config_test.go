package bitxid

import (
	"testing"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	loggerAccountDID = "account-did"
	loggerChainDID   = "chain-did"
)

var w *loggerWrapper

type loggerWrapper struct {
	loggers map[string]*logrus.Entry
}

func testUnmarshalConfig(t *testing.T) {
	config, err := UnmarshalConfig("./config")
	assert.Nil(t, err)
	dConfig, err := DefaultBitXIDConfig()
	assert.Equal(t, *dConfig, config.BitXIDConfig)
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

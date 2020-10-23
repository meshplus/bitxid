package bitxid

import (
	"testing"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	loggerDID    = "did"
	loggerMethod = "method"
)

var w *loggerWrapper

type loggerWrapper struct {
	loggers map[string]*logrus.Entry
}

func TestUnmarshalConfig(t *testing.T) {
	config, err := UnmarshalConfig("./config")
	assert.Nil(t, err)
	dConfig, err := DefaultBitXIDConfig()
	assert.Equal(t, *dConfig, config.BitXIDConfig)
}

// Init initializes all loggers
func loggerInit() {
	// config *repo.Config
	m := make(map[string]*logrus.Entry)
	m[loggerDID] = log.NewWithModule(loggerDID)
	m[loggerDID].Logger.SetLevel(log.ParseLevel("info"))
	m[loggerMethod] = log.NewWithModule(loggerMethod)
	m[loggerMethod].Logger.SetLevel(log.ParseLevel("info"))

	w = &loggerWrapper{loggers: m}
}

// Get gets logger for specific module
func loggerGet(name string) logrus.FieldLogger {
	return w.loggers[name]
}

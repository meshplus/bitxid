package loggers

import (
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/sirupsen/logrus"
)

const (
	DID    = "did"
	Method = "method"
	App    = "app"
	API    = "api"
)

var w *loggerWrapper

type loggerWrapper struct {
	loggers map[string]*logrus.Entry
}

// Init initializes all loggers
func Init() {
	// config *repo.Config
	m := make(map[string]*logrus.Entry)
	m[DID] = log.NewWithModule(DID)
	// m[DID].Logger.SetLevel(log.ParseLevel(config.Log.Module.DID))
	m[DID].Logger.SetLevel(log.ParseLevel("info"))
	m[Method] = log.NewWithModule(Method)
	// m[Method].Logger.SetLevel(log.ParseLevel(config.Log.Module.Method))
	m[Method].Logger.SetLevel(log.ParseLevel("info"))
	m[App] = log.NewWithModule(App)
	// m[App].Logger.SetLevel(log.ParseLevel(config.Log.Level))
	m[App].Logger.SetLevel(log.ParseLevel("info"))
	m[API] = log.NewWithModule(API)
	// m[API].Logger.SetLevel(log.ParseLevel(config.Log.Level))
	m[API].Logger.SetLevel(log.ParseLevel("info"))

	w = &loggerWrapper{loggers: m}
}

// Get gets logger for specific module
func Get(name string) logrus.FieldLogger {
	return w.loggers[name]
}

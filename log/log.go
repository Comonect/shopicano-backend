package log

import (
	"github.com/shopicano/shopicano-backend/log/hooks"
	"github.com/sirupsen/logrus"
	"os"
)

var defLogger *logrus.Logger

func SetupLog() {
	defLogger = logrus.New()
	defLogger.Out = os.Stdout
	defLogger.AddHook(hooks.NewHook())
}

func Log() *logrus.Logger {
	return defLogger
}

package logger

import (
	"github.com/op/go-logging"
	"os"
)

// Logger logger
var Logger *logging.Logger

func init() {
	Logger = logging.MustGetLogger("gortal")
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	stdoutHandler := logging.NewBackendFormatter(logging.NewLogBackend(os.Stdout, "", 0), format)
	logging.SetBackend(stdoutHandler)
}

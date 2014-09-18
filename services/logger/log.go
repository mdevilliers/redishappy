package logger

import (
	"github.com/blackjack/syslog"
	"github.com/natefinch/lumberjack"
	"io"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLogging(logPath string) {
	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslogWriter := &syslog.Writer{LogPriority: syslog.LOG_INFO}

	alloutputs := io.MultiWriter(&lumberjack.Logger{
		Dir:        logPath,
		NameFormat: "redis-happy.log",
		MaxSize:    lumberjack.Gigabyte,
		MaxBackups: 3,
		MaxAge:     28,
	}, os.Stdout, syslogWriter)

	Trace = log.New(alloutputs, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(alloutputs, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(alloutputs, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(alloutputs, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

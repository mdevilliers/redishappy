package logger

import (
	"io"
	"log"
	"os"

	"github.com/blackjack/syslog"
)

func InitLogging(logPath string) {

	syslog.Openlog("redis-happy", syslog.LOG_PID, syslog.LOG_USER)
	syslogWriter := &syslog.Writer{LogPriority: syslog.LOG_ERR}

	logFileWriter := newLogFileWriter(logPath)

	allOutputs := io.MultiWriter(logFileWriter, os.Stdout, syslogWriter)

	Trace = log.New(logFileWriter, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(logFileWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(allOutputs, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(allOutputs, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	NoteWorthy = log.New(allOutputs, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

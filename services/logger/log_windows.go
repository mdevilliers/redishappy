package logger

import (
	"io"
	"log"
	"os"

	"github.com/natefinch/lumberjack"
)

func InitLogging(logPath string) {

	logFileWriter := &lumberjack.Logger{
		Dir:        logPath,
		NameFormat: "redis-happy.log",
		MaxSize:    lumberjack.Megabyte,
		MaxBackups: 3,
		MaxAge:     28,
	}

	allOutputs := io.MultiWriter(logFileWriter, os.Stdout)

	Trace = log.New(allOutputs, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(allOutputs, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(allOutputs, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(allOutputs, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

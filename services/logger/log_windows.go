package logger

import (
	"io"
	"log"
	"os"
)

func InitLogging(logPath string) {

	allOutputs := io.MultiWriter(newLogFileWriter(logPath), os.Stdout)

	Trace = log.New(allOutputs, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(allOutputs, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(allOutputs, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(allOutputs, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	NoteWorthy = log.New(allOutputs, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

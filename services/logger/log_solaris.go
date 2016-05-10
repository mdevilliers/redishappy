package logger

// InitLogging is a NOOP for solaris and will default to StdOut
func InitLogging(logPath string) {}

package log

type devNullLogger struct{}

func NewDevNullLogger() ILogger {
	return &devNullLogger{}
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *devNullLogger) Debug(v ...interface{}) {
}

// Debugf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *devNullLogger) Debugf(format string, v ...interface{}) {
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *devNullLogger) Info(v ...interface{}) {
}

// Infof calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *devNullLogger) Infof(format string, v ...interface{}) {
}

// Warn calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *devNullLogger) Warning(v ...interface{}) {
}

// Warnf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *devNullLogger) Warningf(format string, v ...interface{}) {
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *devNullLogger) Error(v ...interface{}) {
}

// Errorf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *devNullLogger) Errorf(format string, v ...interface{}) {
}

// ErrorIfNotNil calls l.Output to print to the logger if err is not nil.
// Arguments are handled in the manner of fmt.Printf.
func (l *devNullLogger) ErrorIfNotNil(message string, err error) {
}

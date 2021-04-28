package httpclient_test

type logMock struct {
	Format string
	Type   string
	Values []interface{}
}

func (l *logMock) save(tp, format string, values ...interface{}) {
	l.Type = tp
	l.Format = format
	l.Values = values
}

func (l *logMock) Errorf(format string, values ...interface{}) {
	l.save("error", format, values)
}

func (l *logMock) Warnf(format string, values ...interface{}) {
	l.save("warn", format, values)
}

func (l *logMock) Debugf(format string, values ...interface{}) {
	l.save("debug", format, values)
}

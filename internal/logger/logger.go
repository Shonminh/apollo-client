package logger

type LogFunc func(format string, v ...interface{})

func dummyLog(format string, v ...interface{}) {}

var LogDebug LogFunc = dummyLog
var LogInfo LogFunc = dummyLog
var LogError LogFunc = dummyLog

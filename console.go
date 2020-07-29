package glog

import (
	"fmt"
	"os"
	"time"
)

// ConsoleLogger 日志结构体
type ConsoleLogger struct {
	Level LogLevel
}

// NewLog 构造函数
func NewLog(levelStr string) ConsoleLogger {
	level, err := parseString2LogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	return ConsoleLogger{
		Level: level,
	}
}

// enable 是否开启该级别日志
func (c ConsoleLogger) enable(level LogLevel) bool {
	return level >= c.Level
}

// Trace trace level
func (c ConsoleLogger) Trace(format string, a ...interface{}) {
	c.log(Trace, format, a...)
}

// Debug debug level
func (c ConsoleLogger) Debug(format string, a ...interface{}) {
	c.log(Debug, format, a...)
}

// Info info level
func (c ConsoleLogger) Info(format string, a ...interface{}) {
	c.log(Info, format, a...)
}

// Warn info level
func (c ConsoleLogger) Warn(format string, a ...interface{}) {
	c.log(Warn, format, a...)
}

// Error info level
func (c ConsoleLogger) Error(format string, a ...interface{}) {
	c.log(Error, format, a...)
}

// log 控制台记录日志的核心方法
func (c ConsoleLogger) log(level LogLevel, format string, a ...interface{}) {
	if !c.enable(level) {
		return
	}
	msg := fmt.Sprintf(format, a...)
	now := time.Now()
	levelStr, _ := parseLogLevel2String(level)
	formatLen := 5 - len(levelStr)
	for i := 0; i < formatLen; i++ {
		levelStr = " " + levelStr
	}
	funcName, fileName, lineNo := getInfo(3)
	fmt.Fprintf(os.Stdout, "[%s] [%s] [%s:%s:%d] - %s\n", now.Format("2006-01-02 15:04:05"), levelStr, fileName, funcName, lineNo, msg)
}

package glog

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
)

const (
	// UNKNOW 未知，初始化为 0
	UNKNOW LogLevel = iota
	// Trace 初始化为 1
	Trace
	// Debug 初始化为 2
	Debug
	// Info 初始化为 3
	Info
	// Warn 初始化为 4
	Warn
	// Error 初始化为 5
	Error
)

// Logger 接口
type Logger interface{}

// LogLevel 日志级别
type LogLevel uint16

// parseString2LogLevel 将字符串转为 LogLevel
func parseString2LogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "TRACE":
		return Trace, nil
	case "DEBUG":
		return Debug, nil
	case "INFO":
		return Info, nil
	case "WARN":
		return Warn, nil
	case "ERROR":
		return Error, nil
	default:
		err := errors.New("法将字符串转为 LogLevel")
		return UNKNOW, err
	}
}

// parseLogLevel2String 将 LogLevel 转为字符串
func parseLogLevel2String(logLevel LogLevel) (string, error) {
	switch logLevel {
	case Trace:
		return "TRACE", nil
	case Debug:
		return "DEBUG", nil
	case Info:
		return "INFO", nil
	case Warn:
		return "WARN", nil
	case Error:
		return "ERROR", nil
	default:
		err := errors.New("无法将 LogLevel 转为字符串")
		return "", err
	}
}

// getInfo 获取日志调用者的文件名、方法名、行号等信息
func getInfo(skip int) (funcName, fileName string, lineNo int) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		fmt.Println("runtime.Caller error.")
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]
	fileName = path.Base(file)
	return
}

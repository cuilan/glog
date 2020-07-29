package glog

import (
	"fmt"
	"os"
	"path"
	"time"
)

// 往文件中写日志

const (
	// InfoLogSubfix info 级别的日志文件后缀名
	InfoLogSubfix = ".log"
	// ErrorLogSubfix error 级别的日志文件后缀名
	ErrorLogSubfix = "-error.log"
)

// FileLogger 文件日志结构体
type FileLogger struct {
	// 日志级别
	Level LogLevel
	// 日志存放路径
	filePath string
	// 日志文件名称
	fileName string
	// 最大文件大小
	maxFileSize int64
	// 日志文件对象
	fileObj *os.File
	// 错误日志文件对象
	errFileObj *os.File
}

// NewFileLogger 构造函数
func NewFileLogger(levelStr, fp, fn string, maxSize int64) *FileLogger {
	logLevel, err := parseString2LogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	file := &FileLogger{
		Level:       logLevel,
		filePath:    fp,
		fileName:    fn,
		maxFileSize: maxSize,
	}
	err = file.initFile()
	if err != nil {
		panic(err)
	}
	err = file.initErrFile()
	if err != nil {
		panic(err)
	}
	return file
}

func (f *FileLogger) initFile() error {
	fullFileName := path.Join(f.filePath, f.fileName)
	// 打开info日志文件
	fileObj, err := os.OpenFile(fullFileName+InfoLogSubfix, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("open log file error, %v\n", err)
		return err
	}
	f.fileObj = fileObj
	return nil
}

func (f *FileLogger) initErrFile() error {
	fullFileName := path.Join(f.filePath, f.fileName)
	// 打开err日志文件
	errFileObj, err := os.OpenFile(fullFileName+ErrorLogSubfix, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("open err log file error, %v\n", err)
		return err
	}
	f.errFileObj = errFileObj
	return nil
}

// enable 是否开启该级别日志
func (f *FileLogger) enable(level LogLevel) bool {
	return level >= f.Level
}

// needCut 检查文件大小
func (f *FileLogger) needCut(file *os.File) bool {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info error: %v\n", err)
		return false
	}
	// 如果当前文件大小大于最大值
	return fileInfo.Size() >= f.maxFileSize
}

// splitFile 切割文件
func (f *FileLogger) splitFile(file *os.File, level LogLevel) {
	// 关闭当前文件
	file.Close()
	// 重命名原日志文件
	fullFileName := path.Join(f.filePath, f.fileName)
	nowStr := time.Now().Format("20060102150405")
	var newLogName string
	if level >= Error {
		newLogName = fmt.Sprintf("%s-error-%s.log", fullFileName, nowStr)
		os.Rename(fullFileName+ErrorLogSubfix, newLogName)
		err := f.initErrFile()
		if err != nil {
			panic(err)
		}
	} else {
		newLogName = fmt.Sprintf("%s-%s.log", fullFileName, nowStr)
		os.Rename(fullFileName+InfoLogSubfix, newLogName)
		err := f.initFile()
		if err != nil {
			panic(err)
		}
	}
}

// Trace trace level
func (f *FileLogger) Trace(format string, a ...interface{}) {
	f.log(Trace, format, a...)
}

// Debug debug level
func (f *FileLogger) Debug(format string, a ...interface{}) {
	f.log(Debug, format, a...)
}

// Info info level
func (f *FileLogger) Info(format string, a ...interface{}) {
	f.log(Info, format, a...)
}

// Warn info level
func (f *FileLogger) Warn(format string, a ...interface{}) {
	f.log(Warn, format, a...)
}

// Error info level
func (f *FileLogger) Error(format string, a ...interface{}) {
	f.log(Error, format, a...)
}

// log 文件记录日志的核心方法
func (f *FileLogger) log(level LogLevel, format string, a ...interface{}) {
	if !f.enable(level) {
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
	if level >= Error {
		if f.needCut(f.errFileObj) {
			f.splitFile(f.errFileObj, level)
		}
		fmt.Fprintf(f.errFileObj, "[%s] [%s] [%s:%s:%d] - %s\n", now.Format("2006-01-02 15:04:05"), levelStr, fileName, funcName, lineNo, msg)
	} else {
		if f.needCut(f.fileObj) {
			f.splitFile(f.fileObj, level)
		}
		fmt.Fprintf(f.fileObj, "[%s] [%s] [%s:%s:%d] - %s\n", now.Format("2006-01-02 15:04:05"), levelStr, fileName, funcName, lineNo, msg)
	}
}

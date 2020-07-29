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
	// 日志通道
	logChan chan *logMsg
}

// logMsg 日志消息结构体
type logMsg struct {
	Level     LogLevel
	msg       string
	fileName  string
	funcName  string
	timestamp string
	line      int
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
		logChan:     make(chan *logMsg, 50000),
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

	// 开启一个goroutine写日志，开启多个会出现并发问题
	go f.writeLogBackground()
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

	// 开启一个goroutine写日志，开启多个会出现并发问题
	go f.writeLogBackground()
	return nil
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
	funcName, fileName, lineNo := getInfo(3)

	// 将日志发送到通道中
	logMsgPoint := &logMsg{
		Level:     level,
		msg:       msg,
		fileName:  fileName,
		funcName:  funcName,
		timestamp: now.Format("2006-01-02 15:04:05"),
		line:      lineNo,
	}

	// 为防止通道中被写满，造成阻塞，使用select多路复用
	select {
	case f.logChan <- logMsgPoint:
	default:
		// 默认不处理，保证主业务正常处理
	}

}

// writeLogBackground 线程在后台写日志
func (f *FileLogger) writeLogBackground() {
	for {
		select {
		// 从通道中取值
		case logMsgPoint := <-f.logChan:
			level := logMsgPoint.Level

			// 格式化日志级别字符对齐：[DEBUG] [ INFO]
			levelStr, _ := parseLogLevel2String(level)
			formatLen := 5 - len(levelStr)
			for i := 0; i < formatLen; i++ {
				levelStr = " " + levelStr
			}

			// 格式化写入日志格式
			logStringFormat := fmt.Sprintf("[%s] [%s] [%s:%s:%d] - %s\n",
				logMsgPoint.timestamp, levelStr, logMsgPoint.fileName,
				logMsgPoint.funcName, logMsgPoint.line, logMsgPoint.msg)

			if level >= Error {
				// 写入错误日志
				if f.needCut(f.errFileObj) {
					f.splitFile(f.errFileObj, level)
				}
				fmt.Fprintf(f.errFileObj, logStringFormat)
			} else {
				// 写入日志
				if f.needCut(f.fileObj) {
					f.splitFile(f.fileObj, level)
				}
				fmt.Fprintf(f.fileObj, logStringFormat)
			}
		default:
			// 通道中取不到值，先睡眠500毫秒
			time.Sleep(time.Millisecond * 500)
		}
	}
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

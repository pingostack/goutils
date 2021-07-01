package goutils

import (
	"errors"
	"fmt"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type LoggerOptions struct {
	Path  string
	Name  string
	Level string
}

type Logger struct {
	logger *logrus.Logger
}

var (
	defaultLogger *Logger
	logLevelMap   = map[string]logrus.Level{
		"panic": logrus.PanicLevel,
		"fatal": logrus.FatalLevel,
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"trace": logrus.TraceLevel,
	}
)

func NewLogger(opt LoggerOptions) (*Logger, error) {
	if _, ok := logLevelMap[opt.Level]; !ok {
		fmt.Printf("unknown log level: %s", opt.Level)
		return nil, errors.New("unknown log level")
	}

	logger := &Logger{}

	if defaultLogger == nil {
		defaultLogger = logger
	}

	logger.logger = logrus.New()
	logfile := path.Join(opt.Path, opt.Name)

	logWriter, _ := rotatelogs.New(
		// 分割后的文件名称
		logfile+".%Y%m%d.log",
		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(logfile),
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	// accesslogger.SetOutput(logWriter)
	lfHook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logLevel := logLevelMap[opt.Level]
	logger.logger.SetLevel(logLevel)
	// 新增 Hook
	logger.logger.AddHook(lfHook)

	return logger, nil
}

//Fatal Fatal
func (logger *Logger) Fatal(format string, args ...interface{}) {
	logger.logger.Infof(format, args...)
}

//Error Error
func (logger *Logger) Error(format string, args ...interface{}) {
	logger.logger.Errorf(format, args...)
}

//Warn Warn
func (logger *Logger) Warn(format string, args ...interface{}) {
	logger.logger.Warnf(format, args...)
}

//Info Info
func (logger *Logger) Info(format string, args ...interface{}) {
	logger.logger.Infof(format, args...)
}

//Debug Debug
func (logger *Logger) Debug(format string, args ...interface{}) {
	logger.logger.Debugf(format, args...)
}

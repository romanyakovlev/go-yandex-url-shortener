package logger

import (
	"log"

	"go.uber.org/zap"
)

type LoggerInterface interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	DPanic(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	DPanicf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Infoln(args ...interface{})
}

type Logger struct {
	zapLogger *zap.SugaredLogger
}

func GetLogger() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	return &Logger{
		zapLogger: logger.Sugar(),
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.zapLogger.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.zapLogger.Info(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.zapLogger.Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.zapLogger.Error(args...)
}

func (l *Logger) DPanic(args ...interface{}) {
	l.zapLogger.DPanic(args...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.zapLogger.Panic(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.zapLogger.Fatal(args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Errorf(template, args...)
}

func (l *Logger) DPanicf(template string, args ...interface{}) {
	l.zapLogger.DPanicf(template, args...)
}

func (l *Logger) Panicf(template string, args ...interface{}) {
	l.zapLogger.Panicf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.zapLogger.Fatalf(template, args...)
}

func (l *Logger) Infoln(args ...interface{}) {
	l.zapLogger.Info(args...)
}

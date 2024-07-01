// Package logger содержит сервис логгирования, который используется во всем приложении.
package logger

import (
	"log"

	"go.uber.org/zap"
)

// LoggerInterface определяет интерфейс для логгера, поддерживающего различные уровни логирования.
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

// Logger представляет собой структуру логгера, использующего zap.SugaredLogger для логирования.
type Logger struct {
	zapLogger *zap.SugaredLogger // zapLogger - это обертка для упрощенного логирования.
}

// GetLogger создает и возвращает новый экземпляр Logger.
func GetLogger() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err) // В случае ошибки при инициализации логгера, приложение завершает работу.
	}
	return &Logger{
		zapLogger: logger.Sugar(), // Использование SugaredLogger для более удобного синтаксиса.
	}
}

// Debug логирует сообщение с уровнем Debug.
func (l *Logger) Debug(args ...interface{}) {
	l.zapLogger.Debug(args...)
}

// Info логирует сообщение с уровнем Info.
func (l *Logger) Info(args ...interface{}) {
	l.zapLogger.Info(args...)
}

// Warn логирует сообщение с уровнем Warn.
func (l *Logger) Warn(args ...interface{}) {
	l.zapLogger.Warn(args...)
}

// Error логирует сообщение с уровнем Error.
func (l *Logger) Error(args ...interface{}) {
	l.zapLogger.Error(args...)
}

// DPanic логирует сообщение с уровнем DPanic. В режиме разработки вызывает панику.
func (l *Logger) DPanic(args ...interface{}) {
	l.zapLogger.DPanic(args...)
}

// Panic логирует сообщение с уровнем Panic и вызывает панику.
func (l *Logger) Panic(args ...interface{}) {
	l.zapLogger.Panic(args...)
}

// Fatal логирует сообщение с уровнем Fatal и завершает работу приложения.
func (l *Logger) Fatal(args ...interface{}) {
	l.zapLogger.Fatal(args...)
}

// Debugf логирует форматированное сообщение с уровнем Debug.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Debugf(template, args...)
}

// Infof логирует форматированное сообщение с уровнем Info.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Infof(template, args...)
}

// Warnf логирует форматированное сообщение с уровнем Warn.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Warnf(template, args...)
}

// Errorf логирует форматированное сообщение с уровнем Error.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Errorf(template, args...)
}

// DPanicf логирует форматированное сообщение с уровнем DPanic. В режиме разработки вызывает панику.
func (l *Logger) DPanicf(template string, args ...interface{}) {
	l.zapLogger.DPanicf(template, args...)
}

// Panicf логирует форматированное сообщение с уровнем Panic и вызывает панику.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.zapLogger.Panicf(template, args...)
}

// Fatalf логирует форматированное сообщение с уровнем Fatal и завершает работу приложения.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.zapLogger.Fatalf(template, args...)
}

// Infoln логирует сообщение с уровнем Info без форматирования.
func (l *Logger) Infoln(args ...interface{}) {
	l.zapLogger.Info(args...)
}

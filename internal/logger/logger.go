package logger

import "go.uber.org/zap"
import "log"

func GetLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}

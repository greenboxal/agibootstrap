package logging

import "go.uber.org/zap"

var rootLogger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = false
	config.DisableCaller = true
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}
	logger, err := config.Build()

	if err != nil {
		panic(err)
	}

	rootLogger = logger
}

func GetRootLogger() *zap.Logger {
	return rootLogger
}

func GetLogger(name string) *zap.SugaredLogger {
	return rootLogger.Sugar().Named(name)
}

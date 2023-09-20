package logging

import (
	"context"
	"os"
	"path"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootLogger *otelzap.Logger

func init() {
	var cores []zapcore.Core

	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = false
	config.DisableCaller = true

	consoleConfig := config
	consoleConfig.Encoding = "console"
	consoleConfig.OutputPaths = []string{"stderr"}
	consoleConfig.ErrorOutputPaths = []string{"stderr"}
	consoleLogger, err := config.Build()

	if err != nil {
		panic(err)
	}

	cores = append(cores, consoleLogger.Core())

	if base := os.Getenv("PSIDB_PIPE_LOGS"); base != "" {
		if err := os.MkdirAll(base, 0755); err != nil {
			panic(err)
		}

		logPath := path.Join(base, "psidb.log")

		pipeConfig := config
		pipeConfig.Encoding = "json"
		pipeConfig.OutputPaths = []string{logPath}
		pipeConfig.ErrorOutputPaths = []string{logPath}
		pipeLogger, err := pipeConfig.Build()

		if err != nil {
			panic(err)
		}

		cores = append(cores, pipeLogger.Core())
	}

	metricsClient, err := statsd.New(os.Getenv("PSIDB_STATSD"))

	if err != nil {
		//panic(err)
	}

	rootMetrics = metricsClient

	teeCore := zapcore.NewTee(cores...)

	rootZap := zap.New(teeCore, zap.AddCallerSkip(1), zap.Development())
	rootLogger = otelzap.New(rootZap)
	otelzap.ReplaceGlobals(rootLogger)
}

func GetRootLogger() *otelzap.Logger {
	return rootLogger
}

func GetZapRootSugar() *zap.SugaredLogger {
	return rootLogger.Logger.Sugar()
}

func GetZapRootLogger() *zap.Logger {
	return rootLogger.Logger
}

func GetRootSugaredLogger(logger *otelzap.Logger) *otelzap.SugaredLogger {
	return logger.Sugar()
}

func GetLogger(name string) *otelzap.SugaredLogger {
	return otelzap.New(rootLogger.Logger.With(zap.String("logger", name))).Sugar()
}

func GetLoggerCtx(ctx context.Context, name string) otelzap.SugaredLoggerWithCtx {
	return otelzap.New(rootLogger.Logger.With(zap.String("logger", name))).Ctx(ctx).Sugar()
}

var rootMetrics *statsd.Client

func Metrics() *statsd.Client {
	return rootMetrics
}

var Module = fx.Module(
	"logging",
	fx.Provide(GetZapRootLogger),
	fx.Provide(GetZapRootSugar),
	fx.Provide(GetRootLogger),
	fx.Provide(GetRootSugaredLogger),
	fx.Provide(Metrics),
)

package build

import "go.uber.org/zap"

type Log struct {
	*zap.SugaredLogger
}

func NewLog(outputFile string) (*Log, error) {
	cfg := zap.NewDevelopmentConfig()

	cfg.OutputPaths = []string{outputFile}
	cfg.ErrorOutputPaths = []string{outputFile}

	l, err := cfg.Build()

	if err != nil {
		return nil, err
	}

	return &Log{
		SugaredLogger: l.Sugar(),
	}, nil
}

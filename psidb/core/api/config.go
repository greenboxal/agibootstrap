package coreapi

import "github.com/alitto/pond"

type Config struct {
	RootUUID string

	DataDir    string
	ProjectDir string

	ListenEndpoint string
	UseTLS         bool

	Workers ExecutorPoolConfig
}

type ExecutorPoolConfig struct {
	MaxWorkers  int
	MaxCapacity int
}

func (wpc *ExecutorPoolConfig) Build() *pond.WorkerPool {
	return pond.New(wpc.MaxWorkers, wpc.MaxCapacity)
}

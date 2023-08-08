package coreapi

type Config struct {
	RootUUID string

	DataDir    string
	ProjectDir string

	ListenEndpoint string
	UseTLS         bool
}

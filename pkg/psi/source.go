package psi

type Language interface {
	Name() string
	Extensions() []string

	CreateSourceFile(fileName string) SourceFile
}

type SourceFile interface {
	Name() string

	Root() Node
	Error() error

	Load() error
	Replace(code string) error

	ToCode(node Node) (string, error)
}

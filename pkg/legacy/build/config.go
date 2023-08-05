package build

import (
	"os"
	"path"
	"path/filepath"
)

type Configuration struct {
	OutputDirectory string
	BuildDirectory  string

	BuildSteps []Step

	MaxSteps  int
	MaxEpochs int
}

func (bd *Configuration) GetOutputPath(p ...string) string {
	return filepath.Join(bd.OutputDirectory, path.Join(p...))
}

func (bd *Configuration) GetBuildPath(p ...string) string {
	return filepath.Join(bd.BuildDirectory, path.Join(p...))
}

func (bd *Configuration) ResolveOutputDir(p ...string) string {
	buildPath := bd.GetOutputPath(p...)

	if err := os.MkdirAll(buildPath, 0755); err != nil {
		panic(err)
	}

	return buildPath
}

func (bd *Configuration) ResolveBuildDir(p ...string) string {
	buildPath := bd.GetBuildPath(p...)

	if err := os.MkdirAll(buildPath, 0755); err != nil {
		panic(err)
	}

	return buildPath
}

func (bd *Configuration) ResolveOutputFile(p ...string) string {
	buildPath := bd.GetOutputPath(p...)
	dirName := filepath.Dir(buildPath)

	if err := os.MkdirAll(dirName, 0755); err != nil {
		panic(err)
	}

	return buildPath
}

func (bd *Configuration) ResolveBuildFile(p ...string) string {
	buildPath := bd.GetBuildPath(p...)
	dirName := filepath.Dir(buildPath)

	if err := os.MkdirAll(dirName, 0755); err != nil {
		panic(err)
	}

	return buildPath
}

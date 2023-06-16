package codex

import (
	"context"

	"golang.org/x/tools/imports"

	"github.com/greenboxal/agibootstrap/pkg/langs/golang"
)

type FixImportsBuildStep struct{}

// FixImportsBuildStep is responsible for fixing all import errors in the project.
// It processes each file in the project and formats the imports using the goimports tool.
func (s *FixImportsBuildStep) Process(ctx context.Context, p *Project) (result BuildStepResult, err error) {
	for _, file := range p.files {
		opt := &imports.Options{
			FormatOnly: false,
			AllErrors:  true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   4,
			Fragment:   false,
		}

		sf, err := p.GetSourceFile(file.Path())

		if err != nil {
			return result, err
		}

		code, err := sf.ToCode(sf.Root().(golang.Node))

		if err != nil {
			return result, err
		}

		newCode, err := imports.Process(file.Path(), []byte(code), opt)

		if err != nil {
			return result, err
		}

		if string(newCode) != code {
			err = sf.Replace(string(newCode))

			if err != nil {
				return result, err
			}

			result.Changes++
		}
	}

	return result, nil
}

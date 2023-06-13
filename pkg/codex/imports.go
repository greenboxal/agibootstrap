package codex

import (
	"os"

	"golang.org/x/tools/imports"

	"github.com/greenboxal/agibootstrap/pkg/io"
)

type FixImportsBuildStep struct{}

func (s *FixImportsBuildStep) Process(p *Project) (result BuildStepResult, err error) {
	for _, file := range p.files {
		opt := &imports.Options{
			FormatOnly: false,
			AllErrors:  true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   4,
			Fragment:   false,
		}

		code, err := os.ReadFile(file)

		if err != nil {
			return result, err
		}

		newCode, err := imports.Process(file, code, opt)

		if err != nil {
			return result, err
		}

		if string(newCode) != string(code) {
			err = io.WriteFile(file, string(newCode))

			if err != nil {
				return result, err
			}

			result.Changes++
		}
	}

	return
}

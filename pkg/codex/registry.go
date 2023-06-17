package codex

import (
	"path/filepath"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type LanguageFactory func(p *Project) psi.Language

type Registry struct {
	project *Project
	langs   map[psi.LanguageID]psi.Language
}

func NewRegistry(project *Project) *Registry {
	r := &Registry{
		project: project,
		langs:   map[psi.LanguageID]psi.Language{},
	}

	for _, factory := range factories {
		l := factory(project)
		r.langs[l.Name()] = l
	}

	return r
}

func (r *Registry) ResolveExtension(fileName string) psi.Language {
	ext := filepath.Ext(fileName)

	for _, l := range r.langs {
		for _, e := range l.Extensions() {
			if ext == e {
				return l
			}
		}
	}

	return nil
}

func (r *Registry) Resolve(language psi.LanguageID) psi.Language {
	return r.langs[language]
}

var factories = map[psi.LanguageID]LanguageFactory{}

func RegisterLanguage(name psi.LanguageID, factory LanguageFactory) {
	factories[name] = factory
}

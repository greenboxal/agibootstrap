package project

import (
	"path/filepath"

	"github.com/greenboxal/agibootstrap/pkg/psi/langs"
)

type LanguageFactory func(p Project) langs.Language

type Registry struct {
	project Project

	langs map[langs.LanguageID]langs.Language
}

func NewRegistry(project Project) *Registry {
	r := &Registry{
		project: project,
		langs:   map[langs.LanguageID]langs.Language{},
	}

	return r
}

func (r *Registry) Register(language langs.Language) {
	r.langs[language.Name()] = language
}

func (r *Registry) ResolveExtension(fileName string) langs.Language {
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

func (r *Registry) Resolve(language langs.LanguageID) langs.Language {
	return r.langs[language]
}

func (r *Registry) GetLanguage(language langs.LanguageID) langs.Language {
	return r.langs[language]
}

var factories = map[langs.LanguageID]LanguageFactory{}

func GetLanguageFactory(name langs.LanguageID) LanguageFactory {
	return factories[name]
}

func RegisterLanguage(name langs.LanguageID, factory LanguageFactory) {
	factories[name] = factory
}

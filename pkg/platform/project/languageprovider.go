package project

import (
	"path/filepath"
	"sync"
)

type LanguageFactory func(p Project) Language

type LanguageProvider struct {
	project Project

	mu    sync.RWMutex
	langs map[LanguageID]Language
}

func NewLanguageProvider(project Project) *LanguageProvider {
	r := &LanguageProvider{
		project: project,
		langs:   map[LanguageID]Language{},
	}

	return r
}

type LanguageWithHook interface {
	Language

	OnEnabled(p Project)
}

func (r *LanguageProvider) Register(language Language) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if l := r.langs[language.Name()]; l != nil {
		if l != language {
			panic("language already registered")
		}

		return
	}

	r.langs[language.Name()] = language

	if l, ok := language.(LanguageWithHook); ok {
		l.OnEnabled(r.project)
	}
}

func (r *LanguageProvider) ResolveExtension(fileName string) Language {
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

func (r *LanguageProvider) Resolve(language LanguageID) Language {
	return r.langs[language]
}

func (r *LanguageProvider) GetLanguage(language LanguageID) Language {
	return r.langs[language]
}

var factories = map[LanguageID]LanguageFactory{}

func GetLanguageFactory(name LanguageID) LanguageFactory {
	return factories[name]
}

func RegisterLanguage(name LanguageID, factory LanguageFactory) {
	factories[name] = factory
}

package filetypes

import (
	"strings"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("filetypes")

type Registry struct {
	mu sync.RWMutex

	byName      map[string]FileType
	byExtension map[string]FileType
}

func NewRegistry() *Registry {
	return &Registry{
		byName:      make(map[string]FileType),
		byExtension: map[string]FileType{},
	}
}

func (r *Registry) Register(fileType FileType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.byName[fileType.GetName()] != nil {
		panic("duplicate file type: " + fileType.GetName())
	}

	r.byName[fileType.GetName()] = fileType

	for _, ext := range fileType.GetExtensions() {
		if r.byExtension[ext] != nil {
			logger.Warn("duplicate file type for extension: " + ext)
		}

		r.byExtension[ext] = fileType
	}
}

func (r *Registry) GetByName(name string) FileType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.byName[name]
}

func (r *Registry) GetForPath(fileName string) FileType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ft := range r.byName {
		for _, ext := range ft.GetExtensions() {
			if strings.HasSuffix(fileName, ext) {
				return ft
			}
		}
	}

	return nil
}

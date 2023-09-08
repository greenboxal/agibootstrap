package vm

import (
	"sync"

	"rogchap.com/v8go"
)

type ModuleSource struct {
	Name   string
	Source string
}

type ModuleCache struct {
	mu      sync.RWMutex
	modules map[ModuleSource]*CachedModule

	vm *Supervisor
}

func NewModuleCache(vm *Supervisor) *ModuleCache {
	return &ModuleCache{
		vm:      vm,
		modules: map[ModuleSource]*CachedModule{},
	}
}

func (mc *ModuleCache) Get(src ModuleSource) (*CachedModule, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if m := mc.modules[src]; m != nil {
		return m, nil
	}

	return nil, nil
}

func (mc *ModuleCache) Add(mod *CachedModule) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.modules[mod.Source()] = mod
}

type CachedModule struct {
	mu sync.RWMutex

	vm *Supervisor

	cached    *v8go.UnboundScript
	codeCache *v8go.CompilerCachedData

	name string
	src  ModuleSource
}

func NewCachedModule(vm *Supervisor, src ModuleSource) *CachedModule {
	return &CachedModule{
		vm:  vm,
		src: src,
	}
}

func (cm *CachedModule) Name() string         { return cm.name }
func (cm *CachedModule) Source() ModuleSource { return cm.src }

func (cm *CachedModule) load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cached == nil {
		var opts v8go.CompileOptions

		opts.Mode = v8go.CompileModeDefault

		cached, err := cm.vm.iso.CompileUnboundScript(cm.src.Source, cm.src.Name, opts)

		if err != nil {
			return err
		}

		cm.cached = cached
		cm.codeCache = cm.cached.CreateCodeCache()
	}

	return nil
}

func (cm *CachedModule) GetCached(iso *Isolate) (*v8go.UnboundScript, error) {
	if cm.cached == nil {
		if err := cm.load(); err != nil {
			return nil, err
		}
	}

	if iso.iso != cm.vm.iso {
		return iso.iso.CompileUnboundScript(cm.src.Source, cm.src.Name, v8go.CompileOptions{
			Mode:       v8go.CompileModeDefault,
			CachedData: cm.codeCache,
		})
	}

	return cm.cached, nil
}

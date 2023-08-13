package vm

import (
	"fmt"
	"sync"

	"rogchap.com/v8go"
)

type ModuleSource struct {
	Name   string
	Source string
}

type CachedModule struct {
	mu sync.RWMutex

	vm *VM

	cached    *v8go.UnboundScript
	codeCache *v8go.CompilerCachedData

	name   string
	origin string
	src    string
}

func NewCachedModule(vm *VM, name string, src ModuleSource) *CachedModule {
	return &CachedModule{
		vm:     vm,
		name:   name,
		origin: src.Name,
		src:    fmt.Sprintf("(function(module, exports, require) {%s\n})", src.Source),
	}
}

func (cm *CachedModule) Name() string { return cm.name }

func (cm *CachedModule) load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cached == nil {
		var opts v8go.CompileOptions

		opts.Mode = v8go.CompileModeDefault

		cached, err := cm.vm.iso.CompileUnboundScript(cm.src, cm.name, opts)

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
		return iso.iso.CompileUnboundScript(cm.src, cm.name, v8go.CompileOptions{
			Mode:       v8go.CompileModeDefault,
			CachedData: cm.codeCache,
		})
	}

	return cm.cached, nil
}

type ModuleCache struct {
	mu      sync.RWMutex
	modules map[string]*CachedModule

	vm *VM
}

func NewModuleCache(vm *VM) *ModuleCache {
	return &ModuleCache{
		vm:      vm,
		modules: map[string]*CachedModule{},
	}
}

func (mc *ModuleCache) Get(name string) (*CachedModule, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if m := mc.modules[name]; m != nil {
		return m, nil
	}

	return nil, nil
}

func (mc *ModuleCache) Add(mod *CachedModule) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.modules[mod.Name()] = mod
}

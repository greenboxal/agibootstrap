package introspectfx

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sync/errgroup"
)

type typeSystemCache struct {
	m          sync.RWMutex
	augmenters []TypeAugmenter

	dirtySet     mapset.Set[reflect.Type]
	builderMap   map[reflect.Type]*TypeBuilder
	reenterLevel atomic.Int64
}

func newTypeSystemCache() *typeSystemCache {
	return &typeSystemCache{
		dirtySet:   mapset.NewSet[reflect.Type](),
		builderMap: map[reflect.Type]*TypeBuilder{},
	}
}

func (c *typeSystemCache) RegisterProperty(typ Type, prop Property) {
	builder, _ := c.getOrCreateBuilder(typ.RuntimeType(), false)

	builder.WithProperty(prop)
}

func (c *typeSystemCache) RegisterAugmenter(aug TypeAugmenter) {
	c.m.Lock()
	defer c.m.Unlock()

	c.augmenters = append(c.augmenters, aug)
}

func (c *typeSystemCache) getOrCreateBuilder(t reflect.Type, lock bool) (result *TypeBuilder, isNew bool) {
	c.m.Lock()
	defer c.m.Unlock()

	if lock {
		defer func() {
			result.Lock()
		}()
	}

	if cached, ok := c.builderMap[t]; ok {
		return cached, false
	}

	name := t.Name()

	if name == "" {
		name = t.Kind().String()
	}

	builder := NewTypeBuilder(name, t)

	c.builderMap[t] = builder
	c.dirtySet.Add(t)

	return builder, true
}

func (c *typeSystemCache) push() {
	c.m.Lock()
	defer c.m.Unlock()

	c.reenterLevel.Add(1)
}

func (c *typeSystemCache) build() {
	targets := c.dirtySet.ToSlice()

	for _, typ := range targets {
		c.dirtySet.Remove(typ)

		builder, _ := c.getOrCreateBuilder(typ, false)

		for _, aug := range c.augmenters {
			if err := aug.AugmentType(builder); err != nil {
				panic(err)
			}
		}

		builder.Build()
	}
}

func (c *typeSystemCache) pop() {
	c.m.Lock()
	level := c.reenterLevel.Add(-1)

	if level > 0 {
		c.m.Unlock()
	} else {
		var wg errgroup.Group

		c.reenterLevel.Add(1)
		defer c.reenterLevel.Add(-1)
		c.m.Unlock()

		for c.dirtySet.Cardinality() > 0 {
			wg.Go(func() (r error) {
				c.reenterLevel.Add(1)
				defer c.reenterLevel.Add(-1)

				defer func() {
					if e := recover(); e != nil {
						if err, ok := e.(error); ok {
							r = err
						} else {
							r = fmt.Errorf("%v", e)
						}
					}
				}()

				c.build()

				return nil
			})

			if err := wg.Wait(); err != nil {
				panic(err)
			}
		}
	}
}

func (c *typeSystemCache) TypeOf(t reflect.Type) Type {
	c.push()
	defer c.pop()

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	builder, _ := c.getOrCreateBuilder(t, false)

	return builder
}

var globalTypeCache = newTypeSystemCache()

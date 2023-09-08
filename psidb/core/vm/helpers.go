package vm

import (
	"context"
	"reflect"

	"rogchap.com/v8go"
)

const handleMagicNumber uint32 = 0x0A55BEEF

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var anyType = reflect.TypeOf((*any)(nil)).Elem()

func MustSet(obj *v8go.Object, key string, value v8go.Valuer) {
	err := obj.Set(key, value)

	if err != nil {
		panic(err)
	}
}

func MustGet(obj *v8go.Object, key string) *v8go.Value {
	value, err := obj.Get(key)

	if err != nil {
		panic(err)
	}

	return value
}

package collectionsfx

import (
	"reflect"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type ReflectedObservableList struct {
	v reflect.Value
}

func (r *ReflectedObservableList) RuntimeElementType() reflect.Type {
	return r.v.Interface().(BasicObservableList).RuntimeElementType()
}

func (r *ReflectedObservableList) AddListener(listener obsfx.InvalidationListener) {
	r.v.Interface().(BasicObservableList).AddListener(listener)
}

func (r *ReflectedObservableList) RemoveListener(listener obsfx.InvalidationListener) {
	r.v.Interface().(BasicObservableList).RemoveListener(listener)
}

func (r *ReflectedObservableList) Iterator() Iterator[any] {
	//TODO implement me
	panic("implement me")
}

func (r *ReflectedObservableList) Get(index int) any {
	args := []reflect.Value{
		reflect.ValueOf(index),
	}

	return int(r.v.MethodByName("Get").Call(args)[0].Int())
}

func (r *ReflectedObservableList) Slice() []any {
	//TODO implement me
	panic("implement me")
}

func (r *ReflectedObservableList) SubSlice(from, to int) []any {
	//TODO implement me
	panic("implement me")
}

func (r *ReflectedObservableList) Len() int {
	return r.v.Interface().(BasicObservableList).Len()
}

func (r *ReflectedObservableList) Contains(value any) bool {
	args := []reflect.Value{
		reflect.ValueOf(value),
	}

	return r.v.MethodByName("Contains").Call(args)[0].Bool()
}

func (r *ReflectedObservableList) IndexOf(value any) int {
	args := []reflect.Value{
		reflect.ValueOf(value),
	}

	return int(r.v.MethodByName("IndexOf").Call(args)[0].Int())
}

func (r *ReflectedObservableList) AddListListener(listener ListListener[any]) {
}

func (r *ReflectedObservableList) RemoveListListener(listener ListListener[any]) {
	//TODO implement me
	panic("implement me")
}

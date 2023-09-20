package typesystem

import (
	"reflect"
	"strconv"
)

type runtimeMethodArgument struct {
	name string
	typ  Type
}

type basicMethod struct {
	declaringType Type
	name          string

	in  []runtimeMethodArgument
	out []runtimeMethodArgument

	variadic bool
}

func (r *basicMethod) Name() string        { return r.name }
func (r *basicMethod) DeclaringType() Type { return r.declaringType }

func (r *basicMethod) NumIn() int { return len(r.in) }

func (r *basicMethod) In(index int) Type {
	if index < 0 || index >= len(r.in) {
		panic("index out of range")
	}

	return r.in[index].typ
}

func (r *basicMethod) NumOut() int {
	return len(r.out)
}

func (r *basicMethod) Out(index int) Type {
	if index < 0 || index >= len(r.out) {
		panic("index out of range")
	}

	return r.out[index].typ
}

func (r *basicMethod) IsVariadic() bool { return r.variadic }

type reflectedMethod struct {
	basicMethod

	m  reflect.Method
	mt Type
}

func newReflectedMethod(declaringType Type, m reflect.Method, mt Type) *reflectedMethod {
	r := &reflectedMethod{
		m:  m,
		mt: mt,
	}

	r.name = m.Name
	r.declaringType = declaringType
	r.variadic = m.Type.IsVariadic()

	for i := 1; i < m.Type.NumIn(); i++ {
		r.in = append(r.in, runtimeMethodArgument{
			name: "arg" + strconv.FormatInt(int64(i), 10),
			typ:  TypeFrom(m.Type.In(i)),
		})
	}

	for i := 0; i < m.Type.NumOut(); i++ {
		r.out = append(r.out, runtimeMethodArgument{
			name: "return" + strconv.FormatInt(int64(i), 10),
			typ:  TypeFrom(m.Type.Out(i)),
		})
	}

	return r
}

func (r *reflectedMethod) Call(receiver Value, args ...Value) ([]Value, error) {
	allArgs := make([]reflect.Value, len(args)+1)
	result := make([]Value, len(r.out))

	allArgs[0] = receiver.Value()

	for i, arg := range args {
		allArgs[i+1] = arg.Value()
	}

	out := r.m.Func.Call(allArgs)

	for i, v := range out {
		result[i] = ValueFrom(v).UncheckedCast(r.out[i].typ)
	}

	return result, nil
}

func (r *reflectedMethod) CallSlice(receiver Value, args ...Value) ([]Value, error) {
	allArgs := make([]reflect.Value, len(args)+1)
	result := make([]Value, len(r.out))

	allArgs[0] = receiver.Value()

	for i, arg := range args {
		allArgs[i+1] = arg.Value()
	}

	out := r.m.Func.CallSlice(allArgs)

	for i, v := range out {
		result[i] = ValueFrom(v).UncheckedCast(r.out[i].typ)
	}

	return result, nil
}

package aipl

import (
	"context"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
	"github.com/pkg/errors"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type VM struct {
	interpreter *prolog.Interpreter

	tx coreapi.Transaction
}

func NewVM(tx coreapi.Transaction) *VM {
	vm := &VM{}
	vm.tx = tx
	vm.interpreter = prolog.New(nil, nil)

	vm.interpreter.Register1(engine.NewAtom("vertex"), vm.predVertex)
	vm.interpreter.Register2(engine.NewAtom("edges"), vm.predEdges)
	vm.interpreter.Register3(engine.NewAtom("edge"), vm.predEdge)

	return vm
}

func (v *VM) predEdge(vm *engine.VM, from, to, key engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
	var fromNode, toNode psi.Node
	var edgeKey psi.EdgeKey

	switch to.(type) {
	case engine.Variable:
		toNode = nil

	case engine.Atom:
	}
	toNode := v.resolveNode(to)
	fromNode := v.resolveNode(to)

	return k(env)
}

func (v *VM) predEdges(vm *engine.VM, from, edgeKey engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
	fromPath, err := psiPath(from)

	if err != nil {
		return engine.Error(err)
	}

	return engine.Delay(func(ctx context.Context) *engine.Promise {
		from, err := v.tx.Resolve(ctx, fromPath)

		if errors.Is(err, psi.ErrNodeNotFound) {
			return engine.Bool(false)
		} else if err != nil {
			return engine.Error(err)
		}

		edges := from.Edges()

		return engine.Repeat(vm, func(env *engine.Env) *engine.Promise {
			if !edges.Next() {
				return engine.Bool(false)
			}

			edge := edges.Value()
			value := engine.NewAtom(edge.Key().GetKey().String())

			return engine.Unify(vm, edgeKey, value, k, env)
		}, env)
	})
}

func (v *VM) predVertex(vm *engine.VM, ref engine.Term, cont engine.Cont, env *engine.Env) *engine.Promise {
	path, err := psiPath(ref)

	if err != nil {
		return engine.Error(err)
	}

	return engine.Delay(func(ctx context.Context) *engine.Promise {
		_, err := v.tx.Resolve(ctx, path)

		if errors.Is(err, psi.ErrNodeNotFound) {
			return engine.Bool(false)
		} else if err != nil {
			return engine.Error(err)
		}

		return engine.Bool(true)
	})
}

func (v *VM) resolveNode(from engine.Term) psi.Node {

}

func psiPath(table engine.Term) (psi.Path, error) {
	switch table := table.(type) {
	case engine.Atom:
		return psi.ParsePath(string(table))
	}
	return psi.Path{}, engine.InstantiationError(nil)
}

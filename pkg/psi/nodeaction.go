package psi

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type NodeAction interface {
	Name() string
	RequestType() typesystem.Type
	ResponseType() typesystem.Type

	Invoke(ctx context.Context, node Node, request any) (any, error)
}

type NodeActionHandler interface {
	Invoke(ctx context.Context, node Node, request any) (any, error)
}

type NodeActionFunc[T Node, Req any, Res any] func(ctx context.Context, node T, request Req) (Res, error)

func (f NodeActionFunc[T, Req, Res]) Invoke(ctx context.Context, node Node, request any) (any, error) {
	var req Req

	if request != nil {
		req = request.(Req)
	}

	return f(ctx, node.(T), req)
}

type NodeActionDefinition struct {
	Name string `json:"name"`

	RequestType  typesystem.Type `json:"request_type"`
	ResponseType typesystem.Type `json:"response_type"`
}

func (d NodeActionDefinition) ValidateImplementation(impl NodeAction) error {
	if d.RequestType != nil || impl.RequestType() != nil {
		requestCompatible := d.RequestType != nil && d.RequestType.AssignableTo(impl.RequestType())

		if !requestCompatible {
			return fmt.Errorf("request type %s is not assignable to %s", impl.RequestType(), d.RequestType)
		}
	}

	if d.ResponseType != nil || impl.ResponseType() != nil {
		responseCompatible := impl.ResponseType() != nil && impl.ResponseType().AssignableTo(d.ResponseType)

		if !responseCompatible {
			return fmt.Errorf("response type %s is not assignable to %s", impl.ResponseType(), d.ResponseType)
		}

		return nil
	}

	return nil
}

type nodeAction struct {
	definition NodeActionDefinition
	handler    NodeActionHandler
}

func (na *nodeAction) Name() string                  { return na.definition.Name }
func (na *nodeAction) RequestType() typesystem.Type  { return na.definition.RequestType }
func (na *nodeAction) ResponseType() typesystem.Type { return na.definition.ResponseType }

func (na *nodeAction) Invoke(ctx context.Context, node Node, request any) (any, error) {
	return na.handler.Invoke(ctx, node, request)
}

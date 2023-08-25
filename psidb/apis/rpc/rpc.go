package rpc

import (
	"context"
	"reflect"

	"github.com/swaggest/jsonrpc"
	"github.com/swaggest/usecase"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	utils "github.com/greenboxal/aip/aip-sdk/pkg/utils"
)

var tracer = otel.Tracer("jsonrpc",
	trace.WithInstrumentationAttributes(semconv.RPCSystemKey.String("jsonrpc")),
)
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

type RpcService struct {
	*jsonrpc.Handler
}

func NewRpcService() *RpcService {
	handler := &jsonrpc.Handler{}

	apiSchema := jsonrpc.OpenAPI{}
	apiSchema.Reflector().SpecEns().Info.Title = "psidb"
	apiSchema.Reflector().SpecEns().Info.Version = "v1.0.0"

	apiSchema.Reflector().InterceptDefName(func(t reflect.Type, defaultDefName string) string {
		return utils.GetParsedTypeName(t).NormalizedFullNameWithArguments()
	})

	handler.OpenAPI = &apiSchema
	handler.Validator = &jsonrpc.JSONSchemaValidator{}
	handler.SkipResultValidation = true

	return &RpcService{Handler: handler}
}

func mustRegister(srv *jsonrpc.Handler, interfaceName string, target any) {
	value := reflect.ValueOf(target)
	typ := value.Type()

	for i := 0; i < typ.NumMethod(); i++ {
		var inType reflect.Type
		var outType reflect.Type

		m := typ.Method(i)
		mi := value.Method(m.Index)
		mtyp := mi.Type()

		hasCtx := false
		hasError := false

		if !m.IsExported() {
			continue
		}

		if mtyp.NumIn() == 2 {
			if !mtyp.In(0).ConvertibleTo(contextType) {
				continue
			}

			hasCtx = true
			inType = mtyp.In(1)
		} else if mtyp.NumIn() == 1 {
			inType = mtyp.In(0)
		} else {
			continue
		}

		if mtyp.NumOut() == 2 {
			if !mtyp.Out(1).ConvertibleTo(errorType) {
				continue
			}

			hasError = true
			outType = mtyp.Out(0)
		} else if mtyp.NumOut() == 1 {
			outType = mtyp.Out(0)
		} else {
			continue
		}

		if inType == nil {
			inType = reflect.TypeOf(struct{}{})
		}

		if outType == nil {
			outType = reflect.TypeOf(struct{}{})
		}

		for inType.Kind() == reflect.Ptr {
			inType = inType.Elem()
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if inType.Kind() == reflect.Interface {
			continue
		}

		if outType.Kind() == reflect.Interface {
			continue
		}

		operationName := interfaceName + "." + m.Name

		u := usecase.NewIOI(
			reflect.New(inType).Interface(),
			reflect.New(outType).Interface(),
			func(ctx context.Context, input, output interface{}) error {
				ctx, span := tracer.Start(ctx, operationName)
				span.SetAttributes(semconv.RPCService(interfaceName))
				span.SetAttributes(semconv.RPCMethod(m.Name))
				defer span.End()

				var args [2]reflect.Value

				if hasCtx {
					args[0] = reflect.ValueOf(ctx)
					args[1] = reflect.ValueOf(input)
				} else {
					args[0] = reflect.ValueOf(input)
				}

				result := mi.Call(args[:mtyp.NumIn()])

				if hasError && result[1].IsValid() {
					err := result[1].Interface()

					if err != nil {
						return err.(error)
					}
				}

				if len(result) > 0 {
					if outType == errorType {
						if result[0].IsNil() {
							return nil
						}

						return result[0].Interface().(error)
					} else {
						v := result[0]

						if v.IsValid() {
							for v.Kind() == reflect.Ptr {
								v = v.Elem()
							}

							reflect.ValueOf(output).Elem().Set(v)
						}
					}
				}

				return nil
			},
		)

		u.SetName(operationName)

		srv.Add(u)
	}
}

type RpcServiceBinding interface {
	Name() string
	Bind(server *RpcService)
	Implementation() any
}

type rpcServiceBinding[T any] struct {
	name    string
	handler T
}

func (r *rpcServiceBinding[T]) Name() string {
	return r.name
}

func (r *rpcServiceBinding[T]) Implementation() any {
	return r.handler
}

func (r *rpcServiceBinding[T]) Bind(server *RpcService) {
	mustRegister(server.Handler, r.name, r.handler)
}

func BindRpcService[T any](name string) fx.Option {
	return utils.WithBinding[RpcServiceBinding]("rpc-service-bindings", func(handler T) RpcServiceBinding {
		return &rpcServiceBinding[T]{
			name:    name,
			handler: handler,
		}
	})
}

func ProvideRpcService[T any](constructor any, name string) fx.Option {
	return fx.Options(
		fx.Provide(constructor),

		BindRpcService[T](name),
	)
}

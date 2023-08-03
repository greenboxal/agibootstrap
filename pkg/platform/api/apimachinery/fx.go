package apimachinery

import (
	"net/http"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"apimachinery",

	fx.Provide(NewServer),
)

func RegisterHttpService[T http.Handler](path string) fx.Option {
	return fx.Invoke(func(server *Server, handler T) {
		server.Mount(path, handler)
	})
}

func ProvideHttpService[T http.Handler](path string, constructor any) fx.Option {
	return fx.Options(
		fx.Provide(constructor),

		RegisterHttpService[T](path),
	)
}

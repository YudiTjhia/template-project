package interfaces

import (
	"net/http"
)

// IRouteRegistrar is implemented by Chi routers and keeps metadata routes
// independent of a concrete router implementation.
type IRouteRegistrar interface {
	Method(method, pattern string, handler http.Handler)
}

// ICRUDRouteOperations supplies model-specific behavior to generic CRUD routes.
type ICRUDRouteOperations interface {
	List(*http.Request) (any, error)
	Create(*http.Request, []byte) (any, error)
	FindByID(*http.Request, int64) (any, error)
	Update(*http.Request, int64, []byte) error
	Delete(*http.Request, int64) error
}

// IListRouteOperations is used when a resource only exposes a standard list
// endpoint and keeps its remaining endpoints domain-specific.
type IListRouteOperations interface {
	List(*http.Request) (any, error)
}

// IRouteError lets an operation specify the HTTP status for a returned error.
type IRouteError interface {
	error
	HTTPStatus() int
}

type IMRoutes interface {
	RegisterRoutes(router IRouteRegistrar, basePath string, operations ICRUDRouteOperations) error
	RegisterItemRoutes(router IRouteRegistrar, basePath string, operations ICRUDRouteOperations) error
	RegisterListRoute(router IRouteRegistrar, basePath string, operations IListRouteOperations) error
}

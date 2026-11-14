// router/router.go
package router

import (
	"Grippy/internal/transport/http/middlewares"
	"net/http"
	"strings"
)

type Router struct {
	Prefix      string
	mainRouter  *MainRouter
	middlewares []middlewares.Middleware
}

func New(prefix string, mainRouter *MainRouter) *Router {
	prefixNew := prefix
	if !strings.HasPrefix(prefixNew, "/") {
		prefixNew = "/" + prefixNew
	}
	return &Router{
		Prefix:      prefixNew,
		mainRouter:  mainRouter,
		middlewares: make([]middlewares.Middleware, 0),
	}
}

func (r *Router) Use(mw middlewares.Middleware) *Router {
	r.middlewares = append(r.middlewares, mw)
	return r
}

func (r *Router) wrap(h http.Handler, localMW ...middlewares.Middleware) http.Handler {
	var final http.Handler = h

	allMW := make([]middlewares.Middleware, 0)
	allMW = append(allMW, localMW...)                  // Local middleware
	allMW = append(allMW, r.middlewares...)            // Router middleware
	allMW = append(allMW, r.mainRouter.middlewares...) // Global middleware

	// Apply in reverse order: from last to first
	for i := len(allMW) - 1; i >= 0; i-- {
		final = allMW[i](final)
	}
	return final
}

func (r *Router) Handle(path string, h http.Handler, mw ...middlewares.Middleware) {
	r.handle("", path, h, mw...)
}

func (r *Router) handle(method, path string, h http.Handler, mw ...middlewares.Middleware) {
	pattern := r.Prefix + path
	if method != "" {
		pattern = method + " " + pattern
	}
	r.mainRouter.mux.Handle(pattern, r.wrap(h, mw...))
}

func (r *Router) Get(path string, h http.HandlerFunc, mw ...middlewares.Middleware) {
	r.handle(http.MethodGet, path, http.HandlerFunc(h), mw...)
}

func (r *Router) Post(path string, h http.HandlerFunc, mw ...middlewares.Middleware) {
	r.handle(http.MethodPost, path, http.HandlerFunc(h), mw...)
}

func (r *Router) Put(path string, h http.HandlerFunc, mw ...middlewares.Middleware) {
	r.handle(http.MethodPut, path, http.HandlerFunc(h), mw...)
}

func (r *Router) Delete(path string, h http.HandlerFunc, mw ...middlewares.Middleware) {
	r.handle(http.MethodDelete, path, http.HandlerFunc(h), mw...)
}

func (r *Router) Group(prefix string) *Router {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	inheritedMW := make([]middlewares.Middleware, len(r.middlewares))
	copy(inheritedMW, r.middlewares)

	return &Router{
		Prefix:      r.Prefix + prefix,
		mainRouter:  r.mainRouter,
		middlewares: inheritedMW,
	}
}

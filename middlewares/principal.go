package middlewares

import (
	"context"
	"github.com/muety/wakapi/models"
	"net/http"
)

const keyPrincipal = "principal"

type PrincipalContainer struct {
	principal *models.User
}

func (c *PrincipalContainer) SetPrincipal(user *models.User) {
	c.principal = user
}

func (c *PrincipalContainer) GetPrincipal() *models.User {
	return c.principal
}

// This middleware is a bit of a dirty workaround to the fact that a http.Request's context
// does not allow to pass values from an inner to an outer middleware. Calling WithContext() on a
// request shallow-copies the whole request itself and therefore, in a chain of handler1(handler2()),
// handler 1 will not have access to values handler 2 writes to its context. In addition, Context.WithValue
// returns a new context with the old context as a parent.
//
// As a concrete example, SentryMiddleware as well as LoggingMiddleware should be quite the outer layers,
// while AuthenticationMiddleware is on the very inside of the chain. However, we still want sentry or the
// logger to have access to the user object populated by the auth. middleware, if present.
//
// This middleware shall be included as the outermost layers and it injects a stateful container that  does
// nothing but conditionally hold a reference to an authenticated user object.
//
// Other reference: https://stackoverflow.com/questions/55972869/send-errors-to-sentry-with-golang-and-mux

type PrincipalMiddleware struct {
	handler http.Handler
}

func NewPrincipalMiddleware() func(handler http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &PrincipalMiddleware{handler: h}
	}
}

func (p *PrincipalMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), keyPrincipal, &PrincipalContainer{})
	p.handler.ServeHTTP(w, r.WithContext(ctx))
}

func SetPrincipal(r *http.Request, user *models.User) {
	if p := r.Context().Value(keyPrincipal); p != nil {
		p.(*PrincipalContainer).SetPrincipal(user)
	}
}

func GetPrincipal(r *http.Request) *models.User {
	if p := r.Context().Value(keyPrincipal); p != nil {
		return p.(*PrincipalContainer).GetPrincipal()
	}
	return nil
}

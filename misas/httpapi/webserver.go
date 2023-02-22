package httpapi

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/morebec/go-errors/errors"
	"net/http"
	"time"
)

// WebServerOption are functions that allow configuring a WebServer instance.
type WebServerOption func(w *WebServer)

// WithPort allows specifying on which port the WebServer should be listening to.
func WithPort(port int64) WebServerOption {
	return WithAddr(fmt.Sprintf(":%d", port))
}

// WithAddr allows specifying on TCP address the WebServer should be listening to.
func WithAddr(addr string) WebServerOption {
	return func(w *WebServer) {
		w.addr = addr
	}
}

// WithMiddleware appends one or more middlewares onto the Router stack.
func WithMiddleware(middlewares ...func(http.Handler) http.Handler) WebServerOption {
	return func(w *WebServer) {
		w.Router.Use(middlewares...)
	}
}

// WithHeartbeat adds the Heartbeat endpoint middleware that is useful for setting up a path like /health that load balancers or
// uptime testing external services can make a request before hitting any routes.
func WithHeartbeat(pattern string) WebServerOption {
	return WithMiddleware(middleware.Heartbeat(pattern))
}

// WithGetEndpoint adds a GET EndpointFunc to the WebServer
func WithGetEndpoint(pattern string, e EndpointFunc) WebServerOption {
	return func(w *WebServer) {
		w.Router.Get(pattern, endpointHandler(e))
	}
}

// WithPostEndpoint adds a POST EndpointFunc to the WebServer
func WithPostEndpoint(pattern string, e EndpointFunc) WebServerOption {
	return func(w *WebServer) {
		w.Router.Post(pattern, endpointHandler(e))
	}
}

func WithGroup(opts ...EndpointGroupOption) WebServerOption {
	return func(w *WebServer) {
		w.Router.Group(func(r chi.Router) {
			for _, opt := range opts {
				opt(r)
			}
		})
	}
}

type EndpointGroupOption func(r chi.Router)

// WithGetGroupEndpoint adds a GET EndpointFunc to an endpoint Group.
func WithGetGroupEndpoint(pattern string, e EndpointFunc) EndpointGroupOption {
	return func(r chi.Router) {
		r.Get(pattern, endpointHandler(e))
	}
}

// WithPostGroupEndpoint adds a Post EndpointFunc to an endpoint Group.
func WithPostGroupEndpoint(pattern string, e EndpointFunc) EndpointGroupOption {
	return func(r chi.Router) {
		r.Post(pattern, endpointHandler(e))
	}
}

// UseGroupMiddleware appends one or more middlewares onto the Endpoint Group's Router stack.
func UseGroupMiddleware(middlewares ...func(http.Handler) http.Handler) EndpointGroupOption {
	return func(r chi.Router) {
		r.Use(middlewares...)
	}
}

// WebServer implementation of a WebServer that can serve as a base to implement HTTP API web servers.
type WebServer struct {
	Router chi.Router
	addr   string
}

func NewWebServer(opts ...WebServerOption) *WebServer {
	ws := &WebServer{
		addr: ":9090",
	}

	ws.Router = chi.NewRouter()

	// Add default WebServer middleware
	ws.Router.Use(
		middleware.RequestID,
		middleware.CleanPath,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(time.Second*60),

		middleware.AllowContentType("application/json"),
		render.SetContentType(render.ContentTypeJSON),
	)

	for _, opt := range opts {
		opt(ws)
	}

	// Add Not Found handler
	ws.Router.NotFound(endpointHandler(func(r *http.Request) EndpointResponse {
		return NewFailureResponse(errors.NotFoundCode, "404 page not found", nil)
	}))

	return ws
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
func (w *WebServer) ListenAndServe() error {
	return http.ListenAndServe(w.addr, w.Router)
}

// EndpointFunc represents an endpoint to handle a request.
type EndpointFunc func(r *http.Request) EndpointResponse

func endpointHandler(e EndpointFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := e(r)

		w.WriteHeader(response.StatusCode)
		for h, hvs := range response.Headers {
			for _, v := range hvs {
				w.Header().Add(h, v)
			}
		}
		render.JSON(w, r, response)
	}
}

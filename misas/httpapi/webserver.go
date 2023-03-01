package httpapi

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/morebec/go-errors/errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// WebServerOption are functions that allow configuring a WebServer instance.
type WebServerOption func(w *WebServer)

// WithPort allows specifying on which port the WebServer should be listening to.
func WithPort(port int) WebServerOption {
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
		w.Router().Use(middlewares...)
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
		w.Router().Get(pattern, endpointHTTPHandler(e))
	}
}

// WithPostEndpoint adds a POST EndpointFunc to the WebServer
func WithPostEndpoint(pattern string, e EndpointFunc) WebServerOption {
	return func(w *WebServer) {
		w.Router().Post(pattern, endpointHTTPHandler(e))
	}
}

func WithGroup(opts ...EndpointGroupOption) WebServerOption {
	return func(w *WebServer) {
		w.Router().Group(func(r chi.Router) {
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
		r.Get(pattern, endpointHTTPHandler(e))
	}
}

// WithPostGroupEndpoint adds a Post EndpointFunc to an endpoint Group.
func WithPostGroupEndpoint(pattern string, e EndpointFunc) EndpointGroupOption {
	return func(r chi.Router) {
		r.Post(pattern, endpointHTTPHandler(e))
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
	*http.Server
	addr string
}

func (ws *WebServer) Router() chi.Router {
	return ws.Server.Handler.(chi.Router)
}

func NewWebServer(opts ...WebServerOption) *WebServer {
	router := chi.NewRouter()
	ws := &WebServer{
		Server: &http.Server{
			Addr:         ":9090",
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 20 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}

	// Add default WebServer middleware
	router.Use(
		middleware.RequestID,
		middleware.CleanPath,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(time.Second*60),

		middleware.AllowContentType("application/json"),
		render.SetContentType(render.ContentTypeJSON),
	)

	// Add Not Found handler
	router.NotFound(endpointHTTPHandler(func(r *EndpointRequest) EndpointResponse {
		return NewFailureResponse(errors.NotFoundCode, "404 page not found", nil)
	}))

	for _, opt := range opts {
		opt(ws)
	}

	return ws
}

// Start helper method that allows starting the server and handling graceful shutdowns.
func (ws *WebServer) Start() error {

	// Channel that listens for requests to stop the server.
	stopServerChan := make(chan bool, 1)

	// Channel that listens for errors from the server.
	errorChan := make(chan error, 1)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	// Start server
	go func() {
		if err := ws.ListenAndServe(); err != nil {
			errorChan <- err
		}
	}()

	select {
	case err := <-errorChan:
		log.Printf("failed starting server: %s \n", err.Error())
		stopServerChan <- true

	case <-shutdown:
		stopServerChan <- true
	}

	// Wait for shutdown
	_ = <-stopServerChan
	// set up a cancellation context if shutting down the server hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ws.SetKeepAlivesEnabled(false)
	if err := ws.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

// EndpointFunc represents an endpoint to handle a request.
type EndpointFunc func(r *EndpointRequest) EndpointResponse

func endpointHTTPHandler(e EndpointFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var response EndpointResponse
		er, err := NewEndpointRequest(r, w)
		if err != nil {
			response = NewErrorResponse(err)
		} else {
			response = e(er)
			w.WriteHeader(response.StatusCode)
			for h, hvs := range response.Headers {
				for _, v := range hvs {
					w.Header().Add(h, v)
				}
			}
		}

		render.JSON(w, r, response)
	}
}

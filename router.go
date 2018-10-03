// Use this package as drop-in replacement for httprouter with hit counter.
// Do not use it in production, as it's slow - using mutex and all.
// Visit /endpoints or /endpoints/unhit to see hit counter.
package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// Router is a http.Handler which can be used to find untested endpoints.
type Router struct {
	router     *httprouter.Router
	hitCounter endpointLookupTable
	m          sync.Mutex
	// httprouter delegates
	NotFound         http.Handler
	MethodNotAllowed http.Handler
	PanicHandler     func(http.ResponseWriter, *http.Request, interface{})
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New()

// New returns a new initialized Router.
func New() *Router {
	return registerPaths(
		&Router{
			router:     httprouter.New(),
			hitCounter: make(endpointLookupTable),
		},
	)
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (r *Router) GET(path string, handle httprouter.Handle) {
	r.Handle("GET", path, handle)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (r *Router) HEAD(path string, handle httprouter.Handle) {
	r.Handle("HEAD", path, handle)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (r *Router) OPTIONS(path string, handle httprouter.Handle) {
	r.Handle("OPTIONS", path, handle)
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (r *Router) POST(path string, handle httprouter.Handle) {
	r.Handle("POST", path, handle)
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (r *Router) PUT(path string, handle httprouter.Handle) {
	r.Handle("PUT", path, handle)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (r *Router) PATCH(path string, handle httprouter.Handle) {
	r.Handle("PATCH", path, handle)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (r *Router) DELETE(path string, handle httprouter.Handle) {
	r.Handle("DELETE", path, handle)
}

// Handle wraps httprouter functionality.
func (r *Router) Handle(method, path string, handle httprouter.Handle) {
	r.registerRoute(method, path, &handle)

	interceptingHandle := func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		defer r.registerHit(&handle)
		handle(w, req, params)
	}

	r.router.Handle(method, path, interceptingHandle)
}

// ServeFiles wraps httprouter functionality.
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.router.ServeFiles(path, root)
}

// ServeHTTP wraps httprouter functionality.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// We have no clue, so ask httprouter to do the actual work.
	r.router.NotFound = r.NotFound
	r.router.MethodNotAllowed = r.MethodNotAllowed
	r.router.PanicHandler = r.PanicHandler
	r.router.ServeHTTP(w, req)
}

// private methods

type endpoint struct {
	Method string
	Path   string
	Hits   int64
}

type endpointLookupTable map[*httprouter.Handle]*endpoint

const (
	endpointsPath      = "/endpoints"
	endpointsUnhitPath = "/endpoints/unhit"
)

func registerPaths(r *Router) *Router {
	r.GET(endpointsPath, r.handleGetEndpoints)
	r.GET(endpointsUnhitPath, r.handleGetEndpointsUnhit)
	return r
}

func (r *Router) filterEndpoints(unhit bool) []*endpoint {
	var endpoints []*endpoint
	for _, endpoint := range r.hitCounter {
		if unhit && endpoint.Hits > 0 {
			continue
		}

		if endpoint.Path == endpointsPath || endpoint.Path == endpointsUnhitPath {
			continue
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func (r *Router) handleGetEndpoints(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	r.writeEndpoints(w, r.filterEndpoints(false))
}

func (r *Router) handleGetEndpointsUnhit(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	r.writeEndpoints(w, r.filterEndpoints(true))
}

func (r *Router) writeEndpoints(w http.ResponseWriter, endpoints []*endpoint) {
	b, err := json.MarshalIndent(endpoints, "  ", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println(err.Error())
	}
}

func (r *Router) registerRoute(method, path string, handle *httprouter.Handle) {
	r.m.Lock()
	defer r.m.Unlock()
	r.hitCounter[handle] = &endpoint{
		Method: method,
		Path:   path,
	}
}

func (r *Router) registerHit(handle *httprouter.Handle) {
	r.m.Lock()
	defer r.m.Unlock()
	r.hitCounter[handle].Hits++
}

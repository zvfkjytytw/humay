package humayhttpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *HTTPServer) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(h.logging)

	// root handler.
	r.Get("/", h.metricsPage)

	// ping handler.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// handlers for application/json content-type.
	r.Group(func(r chi.Router) {
		r.Use(jsonCtx)
		r.Post("/update", h.putJSONValue)
		r.Post("/value", h.getJSONValue)
	})

	// handler for update metric in text/plain content-type.
	r.Route("/update/{metricType}/{metricName}/{metricValue}", func(r chi.Router) {
		r.Use(updateCtx)
		r.Post("/", h.putValue)
		r.Get("/", notImplementedYet)
	})

	// handler for get value of metric in text/plain content-type.
	r.Route("/value/{metricType}/{metricName}", func(r chi.Router) {
		r.Use(valueCtx)
		r.Get("/", h.getValue)
		r.Post("/", notImplementedYet)
	})

	// stubs.
	r.Get("/*", notImplementedYet)
	r.Post("/*", notImplementedYet)
	r.Put("/*", notImplementedYet)

	return r
}

// not implemented handlers.
func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not implemented yet"))
}

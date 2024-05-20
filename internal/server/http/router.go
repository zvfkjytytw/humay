package humayhttpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *HTTPServer) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	// root handler.
	r.Get("/*", notAllowed)

	// ping handler.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// metrics update.
	r.Route("/update", func(r chi.Router) {
		r.Get("/", notAllowed)
		r.Post("/", notAllowed)
		r.Route("/{metricType}/{metricName}/{metricValue}", func(r chi.Router) {
			r.Use(updateCtx)
			r.Post("/", h.putValue)
		})
	})

	// metrics receive.
	r.Route("/value", func(r chi.Router) {
		r.Get("/", notAllowed)
		r.Post("/", notAllowed)
		r.Route("/{metricType}/{metricName}", func(r chi.Router) {
			r.Use(valueCtx)
			r.Get("/", h.getValue)
		})
	})

	return r
}

func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not implemented yet"))
}

func notAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("not allowed"))
}
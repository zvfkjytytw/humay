package humayhttpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	hm "github.com/zvfkjytytw/humay/internal/server/http/middleware"
)

func (h *HTTPServer) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(hm.Compressor())
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(hm.Logging(h.logger))

	// root handler.
	r.Get("/", h.metricsPage)

	// ping handler.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		err := h.storage.CheckDBConnect(r.Context())
		if err != nil {
			h.logger.Sugar().Errorf("absent db connect: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
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

	// handlers for application/json content-type.
	r.Group(func(r chi.Router) {
		r.Use(jsonCtx)
		r.Post("/update", h.putJSONValue)
		r.Post("/value", h.getJSONValue)
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

package humayhttpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	responseData struct {
		statusCode int
		answerSize int
	}

	loggingResponseWriter struct {
		responseWriter http.ResponseWriter
		responseData   *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.responseWriter.Write(b)
	r.responseData.answerSize = size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func (r *loggingResponseWriter) Header() http.Header {
	return r.responseWriter.Header()
}

func (h *HTTPServer) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rID, ok := r.Context().Value(middleware.RequestIDKey).(string)
		if !ok {
			h.logger.Error("undefined request id")
		}
		method := r.Method
		uri := r.URL.Path

		lw := &loggingResponseWriter{
			responseWriter: w,
			responseData:   &responseData{},
		}

		next.ServeHTTP(lw, r)

		rDuration := time.Since(start).Nanoseconds()
		h.logger.Info(
			fmt.Sprintf("Request %v", rID),
			zap.String("Method", method),
			zap.String("URI", uri),
			zap.String("Duration", fmt.Sprintf("%d ns", rDuration)),
			zap.Int("Response Code", lw.responseData.statusCode),
			zap.Int("Response Length", lw.responseData.answerSize),
		)
	})
}

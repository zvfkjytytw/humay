package humayhttpmiddleware

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
		answerBody string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.answerSize = size
	r.responseData.answerBody = string(b)
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rID, ok := r.Context().Value(middleware.RequestIDKey).(string)
			if !ok {
				logger.Error("undefined request id")
			}
			method := r.Method
			uri := r.URL.Path

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}

			next.ServeHTTP(lw, r)

			rDuration := time.Since(start).Nanoseconds()
			logger.Info(
				fmt.Sprintf("Request %v", rID),
				zap.String("Method", method),
				zap.String("URI", uri),
				zap.String("Duration", fmt.Sprintf("%d ns", rDuration)),
				zap.Int("Response Code", lw.responseData.statusCode),
				zap.Int("Response Length", lw.responseData.answerSize),
				// zap.String("Response Body", lw.responseData.answerBody), // for debug
			)
		})
	}
}
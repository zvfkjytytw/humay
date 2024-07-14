package humayhttpmiddleware

import (
	// "bytes"
	"fmt"
	// "io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	responseData struct {
		statusCode int
		answerSize int
		// answerBody string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.answerSize = size
	// r.responseData.answerBody = string(b)
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
			// cType := r.Header.Get("Content-Type")
			// aEnc := r.Header.Get("Accept-Encoding")
			// cEnc := r.Header.Get("Content-Encoding")
			reqHash := r.Header.Get("HashSHA256")

			// // read request body
			// bodyBytes, _ := io.ReadAll(r.Body)
			// requestBody := string(bodyBytes)
			// r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}

			next.ServeHTTP(lw, r)

			respHeaders := lw.Header()
			var respHash, respHashKey string
			values, ok := respHeaders["HashSHA256"]
			if ok {
				respHash = values[0]
			} else {
				respHash = "unknown"
			}

			values, ok = respHeaders["HashKey"]
			if ok {
				respHashKey = values[0]
			} else {
				respHashKey = "unknown"
			}

			rDuration := time.Since(start).Nanoseconds()
			logger.Info(
				fmt.Sprintf("Request %v", rID),
				zap.String("Method", method),
				zap.String("URI", uri),
				// zap.String("Content-Type", cType),      // for debug
				// zap.String("Content-Encodig", cEnc),    // for debug
				// zap.String("Accept-Encodig", aEnc),     // for debug
				// zap.String("RequestBody", requestBody), // for debug
				zap.String("RequestHash", reqHash),         // for debug
				zap.String("ResponsetHash", respHash),      // for debug
				zap.String("ResponseHashKey", respHashKey), // for debug
				zap.String("Duration", fmt.Sprintf("%d ns", rDuration)),
				zap.Int("Response Code", lw.responseData.statusCode),
				zap.Int("Response Length", lw.responseData.answerSize),
				// zap.String("Response Body", lw.responseData.answerBody), // for debug
			)
		})
	}
}

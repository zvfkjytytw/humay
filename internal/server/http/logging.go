package humayhttpserver

import (
	"bytes"
	"fmt"
	"io"
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

func (h *HTTPServer) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rID, ok := r.Context().Value(middleware.RequestIDKey).(string)
		if !ok {
			h.logger.Error("undefined request id")
		}
		method := r.Method
		uri := r.URL.Path
		var b bytes.Buffer
		b.ReadFrom(r.Body)
		r.Body = io.NopCloser(&b)

		var body string
		var rBody []byte

		_, err := b.Read(rBody)
		if err != nil {
			fmt.Printf("read body error: %v\n", err)
			body = "failed read body"
		} else {
			body = string(rBody)
		}

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		next.ServeHTTP(lw, r)

		rDuration := time.Since(start).Nanoseconds()
		h.logger.Info(
			fmt.Sprintf("Request %v", rID),
			zap.String("Method", method),
			zap.String("URI", uri),
			zap.String("ReqBody", body),
			zap.String("Duration", fmt.Sprintf("%d ns", rDuration)),
			zap.Int("Response Code", lw.responseData.statusCode),
			zap.Int("Response Length", lw.responseData.answerSize),
			zap.String("RespBody", lw.responseData.answerBody),
		)
	})
}

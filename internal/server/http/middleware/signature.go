package humayhttpmiddleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	humayCommon "github.com/zvfkjytytw/humay/internal/common"
)

type signingResponseWriter struct {
	http.ResponseWriter
	hashKey string
}

func (s *signingResponseWriter) Write(b []byte) (int, error) {
	size, err := s.ResponseWriter.Write(b)
	if s.hashKey != "" {
		hash := fmt.Sprintf("%x", humayCommon.Hash256(b, s.hashKey))
		s.ResponseWriter.Header().Set("HashSHA256", hash)
	}

	return size, err
}

func Signature(hashKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hashKey != "" && r.URL.Path != "/" {
				requestBodyHash := r.Header.Get("HashSHA256")
				if requestBodyHash == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("absent body hash header"))
					return
				}

				// read request body
				bodyBytes, _ := io.ReadAll(r.Body)
				hash := fmt.Sprintf("%x", humayCommon.Hash256(bodyBytes, hashKey))
				if hash != requestBodyHash {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("hashs not equal"))
					return
				}

				// rewrite body to request
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			sw := &signingResponseWriter{
				ResponseWriter: w,
				hashKey:        hashKey,
			}
			next.ServeHTTP(sw, r)
		})
	}
}

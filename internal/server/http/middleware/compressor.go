package humayhttpmiddleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressedResponseWriter struct {
	http.ResponseWriter
}

func (w *compressedResponseWriter) Write(b []byte) (int, error) {
	writer, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.BestCompression)
	if err != nil {
		w.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		w.ResponseWriter.Write([]byte("failed gzip response body"))
		return 0, err
	}
	defer writer.Close()

	return writer.Write(b)
}

// func Compressor() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			cType := r.Header.Get("Content-Type")
// 			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && (cType == "application/json" || cType == "text/html") {
// 				switch r.Header.Get("Content-Encoding") {
// 				case "gzip":
// 					gz, err := gzip.NewReader(r.Body)
// 					if err != nil {
// 						w.WriteHeader(http.StatusInternalServerError)
// 						w.Write([]byte("failed read compressed body"))
// 						return
// 					}
// 					r.Body = gz
// 					w.Header().Set("Accept-Encoding", "gzip")
// 					w.Header().Set("Content-Encoding", "gzip")
// 					w = &compressedResponseWriter{ResponseWriter: w}
// 				}
// 			}
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

func Compressor() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cType := r.Header.Get("Content-Type")
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && (cType == "application/json" || cType == "text/html") {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("failed read compressed body"))
					return
				}
				r.Body = gz
				w.Header().Set("Accept-Encoding", "gzip")
				w.Header().Set("Content-Encoding", "gzip")
				w = &compressedResponseWriter{ResponseWriter: w}
			}
			next.ServeHTTP(w, r)
		})
	}
}

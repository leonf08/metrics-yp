package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

var contentTypes = []string{"application/json", "html/text"}

type compressWriter struct {
	w   http.ResponseWriter
	gzw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:   w,
		gzw: gzip.NewWriter(w),
	}
}

// Write writes the compressed data to the underlying gzip writer.
func (cw *compressWriter) Write(b []byte) (int, error) {
	return cw.gzw.Write(b)
}

// WriteHeader writes the header to the underlying writer.
func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}

	cw.w.WriteHeader(statusCode)
}

// Header returns the header from the underlying writer.
func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

// Close closes the underlying gzip writer.
func (cw *compressWriter) Close() error {
	return cw.gzw.Close()
}

type compressReader struct {
	r   io.ReadCloser
	gzr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:   r,
		gzr: zr,
	}, nil
}

// Read reads the decompressed data from the underlying gzip reader.
func (cr *compressReader) Read(p []byte) (n int, err error) {
	return cr.gzr.Read(p)
}

// Close closes the underlying gzip reader.
func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.gzr.Close()
}

// Compress is a middleware that compresses the response body
// and decompresses the request body. It uses gzip compression.
// It only compresses the response if the client supports it.
// It only decompresses the request if the client sets the Content-Encoding header to gzip.
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if slices.Contains(contentTypes, r.Header.Get("Accept")) {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
		}

		next.ServeHTTP(ow, r)
	})
}

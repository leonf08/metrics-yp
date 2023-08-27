package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"
)

var contentTypes = []string{"application/json", "text/html"}

type compressWriter struct {
	w http.ResponseWriter
	gzw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
    return &compressWriter{
        w:  w,
        gzw: gzip.NewWriter(w),
    }
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	return cw.gzw.Write(b)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}

	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (c *compressWriter) Close() error {
    return c.gzw.Close()
}

type compressReader struct {
	r io.ReadCloser
	gzr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r: r,
		gzr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
    return c.gzr.Read(p)
}

func (c *compressReader) Close() error {
    if err := c.r.Close(); err != nil {
        return err
    }
    return c.gzr.Close()
}

func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ow := w

        fmt.Println(r.Header.Get("Content-Type"))
        if slices.Contains(contentTypes, r.Header.Get("Content-Type")) {
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
                    ow.WriteHeader(http.StatusInternalServerError)
                    return
                }
                r.Body = cr
                defer cr.Close()
            }
        }

        h.ServeHTTP(ow, r)
	})
}
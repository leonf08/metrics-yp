package httpserver

import (
	"compress/gzip"
	"io"
	"net/http"
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

func (cr compressReader) Read(p []byte) (n int, err error) {
	return cr.gzr.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.gzr.Close()
}

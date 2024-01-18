package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	err    chan error
}

func NewServer(h http.Handler, address string) *Server {
	s := &Server{
		server: &http.Server{
			Addr:    address,
			Handler: h,
		},
		err: make(chan error, 1),
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		s.err <- s.server.ListenAndServe()
		close(s.err)
	}()
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

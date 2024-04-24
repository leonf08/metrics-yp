package grpc

import (
	"net"
	"net/netip"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"github.com/leonf08/metrics-yp.git/internal/proto"
	"github.com/leonf08/metrics-yp.git/internal/server/grpc/interceptors"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Server struct {
	server  *grpc.Server
	repo    repo.Repository
	fs      services.FileStore
	log     zerolog.Logger
	address string
	err     chan error
}

func NewServer(repo repo.Repository, fs services.FileStore, log zerolog.Logger, address, trustedSubnet string) *Server {
	var i []grpc.UnaryServerInterceptor

	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	i = append(i, logging.UnaryServerInterceptor(interceptors.InterceptorLogger(log), logOpts...))

	if trustedSubnet != "" {
		trustedPeers := []netip.Prefix{
			netip.MustParsePrefix(trustedSubnet),
		}

		headers := []string{realip.XRealIp}
		ipOpts := []realip.Option{
			realip.WithTrustedPeers(trustedPeers),
			realip.WithHeaders(headers),
		}

		i = append(i, realip.UnaryServerInterceptorOpts(ipOpts...))
	}

	serverOpt := grpc.ChainUnaryInterceptor(i...)

	sg := grpc.NewServer(serverOpt)

	s := &Server{
		server:  sg,
		repo:    repo,
		fs:      fs,
		log:     log,
		address: address,
		err:     make(chan error, 1),
	}

	go s.start()

	return s
}

func (s *Server) start() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.err <- err
		close(s.err)
		return
	}

	proto.RegisterMetricsServer(s.server, newMetricsServer(s.repo, s.fs, s.log))
	s.err <- s.server.Serve(listener)
	close(s.err)
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}

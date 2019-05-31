package component_tests_test

import (
	"net"

	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MetronServer struct {
	port     int
	server   *grpc.Server
	listener net.Listener
	Metron   *mockMetronIngressServer
}

func NewMetronServer() (*MetronServer, error) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		CAFilePath(),
		MetronCertPath(),
		MetronKeyPath(),
	)
	if err != nil {
		return nil, err
	}
	transportCreds := credentials.NewTLS(tlsConfig)
	mockMetron := newMockMetronIngressServer()

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(grpc.Creds(transportCreds))
	loggregator_v2.RegisterIngressServer(s, mockMetron)

	go s.Serve(lis)

	return &MetronServer{
		port:     lis.Addr().(*net.TCPAddr).Port,
		server:   s,
		listener: lis,
		Metron:   mockMetron,
	}, nil
}

func (s *MetronServer) URI() string {
	return s.listener.Addr().String()
}

func (s *MetronServer) Port() int {
	return s.port
}

func (s *MetronServer) Stop() error {
	err := s.listener.Close()
	s.server.Stop()
	return err
}

package component_tests_test

import (
	"log"
	"net"

	"code.cloudfoundry.org/tlsconfig"
	"code.cloudfoundry.org/tlsconfig/certtest"
	"github.com/cloudfoundry/statsd-injector/component_tests/fakes"

	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MetronServer struct {
	port     int
	server   *grpc.Server
	listener net.Listener
	Metron   *fakes.FakeMetronIngressServer
}

func NewMetronServer(ca *certtest.Authority, caFile string) (*MetronServer, error) {
	certPath, keyPath := GenerateCertKey("metron", ca)
	config, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(certPath, keyPath),
	).Client(
		tlsconfig.WithAuthorityFromFile(caFile),
		tlsconfig.WithServerName("metron"),
	)
	if err != nil {
		log.Fatal("Invalid TLS credentials")
	}

	transportCreds := credentials.NewTLS(config)
	mockMetron := &fakes.FakeMetronIngressServer{}
	mockMetron.SenderStub = func(stream loggregator_v2.Ingress_SenderServer) error {
		for {
			_, err := stream.Recv()
			if err != nil {
				return err
			}
		}
	}

	lis, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(grpc.Creds(transportCreds))
	loggregator_v2.RegisterIngressServer(s, mockMetron)

	go s.Serve(lis) //nolint:errcheck

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

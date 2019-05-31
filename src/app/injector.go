package app

import (
	"fmt"
	"log"

	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/statsd-injector/internal/egress"
	"github.com/cloudfoundry/statsd-injector/internal/ingress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	StatsdHost string
	StatsdPort uint
	MetronPort uint

	CA   string
	Cert string
	Key  string
}

type Injector struct {
	statsdPort uint
	apiVersion string
	metronPort uint

	ca   string
	cert string
	key  string
}

func NewInjector(c Config) *Injector {
	return &Injector{
		statsdPort: c.StatsdPort,
		metronPort: c.MetronPort,
		ca:         c.CA,
		cert:       c.Cert,
		key:        c.Key,
	}
}

func (i *Injector) Start() {
	inputChan := make(chan *loggregator_v2.Envelope)
	hostport := fmt.Sprintf("localhost:%d", i.statsdPort)

	_, addr := ingress.Start(hostport, inputChan)

	log.Printf("Started statsd-injector listener at %s", addr)

	config, err := loggregator.NewIngressTLSConfig(i.ca, i.cert, i.key)
	if err != nil {
		log.Fatal("Invalid TLS credentials")
	}
	statsdEmitter := egress.New(fmt.Sprintf("localhost:%d", i.metronPort),
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
	)
	statsdEmitter.Run(inputChan)
}

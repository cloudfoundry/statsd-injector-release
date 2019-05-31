package egress

import "code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

//go:generate hel

type MetronIngressServer interface {
	loggregator_v2.IngressServer
}

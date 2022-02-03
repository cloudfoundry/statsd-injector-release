package egress

import "code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"

//go:generate hel

type MetronIngressServer interface {
	loggregator_v2.IngressServer
}

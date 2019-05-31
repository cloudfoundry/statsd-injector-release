//go:generate hel
package component_tests

import "code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

type MetronIngressServer interface {
	loggregator_v2.IngressServer
}

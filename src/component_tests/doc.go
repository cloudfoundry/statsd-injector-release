//go:generate hel
package component_tests

import "code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"

type MetronIngressServer interface {
	loggregator_v2.IngressServer
}

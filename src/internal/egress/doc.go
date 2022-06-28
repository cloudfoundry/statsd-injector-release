package egress

import "code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"

//go:generate go run git.sr.ht/~nelsam/hel/v3 -t MetronIngressServer -o helheim_fixed_test.go

type MetronIngressServer interface {
	loggregator_v2.IngressServer
}

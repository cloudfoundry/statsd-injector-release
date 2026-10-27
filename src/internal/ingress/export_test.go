package ingress

// SafeParse exposes the unexported safeParse to the external ingress_test
// package so the panic-recovery path can be tested by injecting a parse func
// that panics. Being a _test.go file, it is not compiled into the binary.
var SafeParse = safeParse

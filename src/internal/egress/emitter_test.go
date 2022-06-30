package egress_test

import (
	"log"
	"net"

	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"github.com/cloudfoundry/statsd-injector/internal/egress"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statsdemitter", func() {
	var (
		serverAddr string
		mockServer *mockMetronIngressServer
		inputChan  chan *loggregator_v2.Envelope
		message    *loggregator_v2.Envelope
	)

	BeforeEach(func() {
		inputChan = make(chan *loggregator_v2.Envelope, 100)
		message = &loggregator_v2.Envelope{
			Message: &loggregator_v2.Envelope_Counter{
				Counter: &loggregator_v2.Counter{
					Name:  "a-name",
					Delta: 48,
				},
			},
		}
	})

	Context("when the server is already listening", func() {
		BeforeEach(func() {
			serverAddr, mockServer = startServer()
			dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
			emitter := egress.New(serverAddr, dialOpt)

			go emitter.Run(inputChan)
		})

		It("emits envelope", func() {
			go keepWriting(inputChan, message)
			var receiver loggregator_v2.Ingress_SenderServer
			Eventually(mockServer.SenderInput.Arg0).Should(Receive(&receiver))

			f := func() bool {
				env, err := receiver.Recv()
				if err != nil {
					return false
				}

				return env.GetCounter().GetDelta() == 48
			}
			Eventually(f).Should(BeTrue())
		})

		It("reconnects when the stream has been closed", func() {
			go keepWriting(inputChan, message)
			close(mockServer.SenderOutput.Ret0)

			f := func() int {
				return len(mockServer.SenderCalled)
			}
			Eventually(f).Should(BeNumerically(">", 1))
		})
	})
})

func keepWriting(c chan<- *loggregator_v2.Envelope, e *loggregator_v2.Envelope) {
	for {
		c <- e
	}
}

func startServer() (string, *mockMetronIngressServer) {
	lis, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	mockMetronIngressServer := newMockMetronIngressServer()
	loggregator_v2.RegisterIngressServer(s, mockMetronIngressServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Failed to start server: %s", err)
		}
	}()

	return lis.Addr().String(), mockMetronIngressServer
}

package egress_test

import (
	"errors"
	"log"
	"net"

	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cloudfoundry/statsd-injector/internal/egress"
	"github.com/cloudfoundry/statsd-injector/internal/egress/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var errStreamClosed = errors.New("stream closed")
var _ = Describe("Statsdemitter", func() {
	var (
		serverAddr string
		mockServer *fakes.FakeMetronIngressServer
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
		var messageChan chan *loggregator_v2.Envelope

		BeforeEach(func() {
			serverAddr, mockServer = startServer()

			messageChan = make(chan *loggregator_v2.Envelope, 100)

			mockServer.SenderStub = func(stream loggregator_v2.Ingress_SenderServer) error {
				for {
					envelope, err := stream.Recv()
					if err != nil {
						return err
					}
					select {
					case messageChan <- envelope:
					default:
					}
				}
			}

			dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
			emitter := egress.New(serverAddr, dialOpt)

			go emitter.Run(inputChan)
		})

		It("emits envelope", func() {
			go keepWriting(inputChan, message)

			var receivedEnvelope *loggregator_v2.Envelope
			Eventually(messageChan).Should(Receive(&receivedEnvelope))

			Expect(receivedEnvelope.GetCounter().GetDelta()).To(Equal(uint64(48)))
		})

		It("reconnects when the stream has been closed", func() {
			go keepWriting(inputChan, message)

			mockServer.SenderReturns(errStreamClosed)

			Eventually(func() int {
				return mockServer.SenderCallCount()
			}).Should(BeNumerically(">", 1))
		})
	})
})

func keepWriting(c chan<- *loggregator_v2.Envelope, e *loggregator_v2.Envelope) {
	for {
		c <- e
	}
}

func startServer() (string, *fakes.FakeMetronIngressServer) {
	lis, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	mockMetronIngressServer := &fakes.FakeMetronIngressServer{}
	loggregator_v2.RegisterIngressServer(s, mockMetronIngressServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Failed to start server: %s", err)
		}
	}()

	return lis.Addr().String(), mockMetronIngressServer
}

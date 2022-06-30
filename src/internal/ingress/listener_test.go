package ingress_test

import (
	"net"

	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"

	"github.com/cloudfoundry/statsd-injector/internal/ingress"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatsdListener", func() {
	Describe("Run", func() {
		var (
			envelopeChan chan *loggregator_v2.Envelope
			listener     *ingress.StatsdListener
			connection   net.Conn
		)

		BeforeEach(func() {
			envelopeChan = make(chan *loggregator_v2.Envelope, 100)

			var addr string
			listener, addr = ingress.Start("localhost:0", envelopeChan)

			var err error
			connection, err = net.Dial("udp4", addr)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			listener.Stop()
			connection.Close()
		})

		It("reads multiple gauges (on different lines) in the same packet", FlakeAttempts(5), func() {
			statsdmsg := []byte("fake-origin.test.gauge:23|g\nfake-origin.other.thing:42|g\nfake-origin.sampled.gauge:17.5|g|@0.2")

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope
			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 23, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 42, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.gauge", 87.5, "gauge")
		})

		It("processes gauge increment/decrement stats", func() {
			statsdmsg := []byte("fake-origin.test.gauge:23|g\nfake-origin.test.gauge:+7|g\nfake-origin.test.gauge:-5|g")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 23, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 30, "gauge")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.gauge", 25, "gauge")
		})

		It("reads multiple timings (on different lines) in the same packet", FlakeAttempts(5), func() {
			statsdmsg := []byte("fake-origin.test.timing:23|ms\nfake-origin.other.thing:420|ms\nfake-origin.sampled.timing:71|ms|@0.1")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.timing", 23, "ms")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 420, "ms")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.timing", 710, "ms")
		})

		It("reads multiple counters (on different lines) in the same packet", FlakeAttempts(5), func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.other.thing:420|c\nfake-origin.sampled.counter:71|c|@0.1")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 23, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "other.thing", 420, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "sampled.counter", 710, "counter")
		})

		It("processes counter increment/decrement stats", func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.test.counter:+7|c\nfake-origin.test.counter:-5|c")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 23, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 30, "counter")

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			checkValueMetric(receivedEnvelope, "fake-origin", "test.counter", 25, "counter")
		})

		It("adds meta-data tags", func() {
			statsdmsg := []byte("fake-origin.test.counter:23|c\nfake-origin.test.counter:+7|c\nfake-origin.test.counter:-5|c")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			Expect(receivedEnvelope.GetTags()["origin"]).To(Equal("fake-origin"))
		})

		It("supports Datadog-style tags", func() {
			statsdmsg := []byte("fake-origin.test.counter:-5|c|#testtag1:testvalue1,testtag2:testvalue2")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			var receivedEnvelope *loggregator_v2.Envelope

			Eventually(envelopeChan).Should(Receive(&receivedEnvelope))
			Expect(receivedEnvelope.GetTags()["testtag1"]).To(Equal("testvalue1"))
			Expect(receivedEnvelope.GetTags()["testtag2"]).To(Equal("testvalue2"))
		})
	})
})

func checkValueMetric(receivedEnvelope *loggregator_v2.Envelope, origin string, name string, value float64, unit string) {
	Expect(receivedEnvelope.GetSourceId()).To(Equal(origin))
	m, ok := receivedEnvelope.GetGauge().GetMetrics()[name]
	Expect(ok).To(BeTrue())
	Expect(m.GetValue()).To(Equal(value))
	Expect(m.GetUnit()).To(Equal(unit))
}

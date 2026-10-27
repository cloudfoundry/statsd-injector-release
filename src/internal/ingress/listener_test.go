package ingress_test

import (
	"net"

	"code.cloudfoundry.org/go-loggregator/v10/rpc/loggregator_v2"

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

		DescribeTable("malformed Datadog tags",
			func(section string, expectedTags map[string]string) {
				statsdmsg := []byte("fake-origin.test.counter:1|c|#" + section)
				_, err := connection.Write(statsdmsg)
				Expect(err).ToNot(HaveOccurred())

				f := func() int {
					connection.Write(statsdmsg) //nolint:errcheck
					return len(envelopeChan)
				}
				Eventually(f, 3).ShouldNot(Equal(0))

				var receivedEnvelope *loggregator_v2.Envelope
				Eventually(envelopeChan).Should(Receive(&receivedEnvelope))

				Expect(receivedEnvelope.GetTags()).To(Equal(expectedTags))
			},
			Entry(nil, "testtag1:testvalue1,testtag2:testvalue2", map[string]string{
				"origin":   "fake-origin",
				"testtag1": "testvalue1",
				"testtag2": "testvalue2",
			}),
			Entry(nil, "x", map[string]string{
				"origin": "fake-origin",
			}),
			Entry(nil, "k:v,bare", map[string]string{
				"origin": "fake-origin",
				"k":      "v",
			}),
			Entry(nil, "bare,k:v", map[string]string{
				"origin": "fake-origin",
				"k":      "v",
			}),
			Entry(nil, "k:v,", map[string]string{
				"origin": "fake-origin",
				"k":      "v",
			}),
			Entry(nil, ",", map[string]string{
				"origin": "fake-origin",
			}),
			Entry(nil, ":v", map[string]string{
				"origin": "fake-origin",
				"":       "v",
			}),
			Entry(nil, "k:", map[string]string{
				"origin": "fake-origin",
				"k":      "",
			}),
			Entry(nil, "a:b:c", map[string]string{
				"origin": "fake-origin",
				"a":      "b",
			}),
		)

		It("keeps processing after a malformed tag line", func() {
			statsdmsg := []byte("fake-origin.test.counter:1|c|#x\nfake-origin.after.gauge:7|g")
			_, err := connection.Write(statsdmsg)
			Expect(err).ToNot(HaveOccurred())

			f := func() int {
				connection.Write(statsdmsg) //nolint:errcheck
				return len(envelopeChan)
			}
			Eventually(f, 3).ShouldNot(Equal(0))

			Eventually(envelopeChan, 3).Should(Receive(WithTransform(func(e *loggregator_v2.Envelope) map[string]*loggregator_v2.GaugeValue { return e.GetGauge().GetMetrics() }, HaveKey("after.gauge"))))
		})

		It("safeParse recovers from a panicking parse func", func() {
			var env *loggregator_v2.Envelope
			var err error

			Expect(func() {
				env, err = ingress.SafeParse(func(string) (*loggregator_v2.Envelope, error) {
					panic("test panic")
				}, "fake-origin.test.counter:1|c")
			}).ToNot(Panic())

			Expect(env).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("panic parsing stat line"))

			expectedEnv := &loggregator_v2.Envelope{SourceId: "fake-origin"}
			env, err = ingress.SafeParse(func(string) (*loggregator_v2.Envelope, error) {
				return expectedEnv, nil
			}, "fake-origin.test.counter:1|c")

			Expect(err).ToNot(HaveOccurred())
			Expect(env).To(BeIdenticalTo(expectedEnv))
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

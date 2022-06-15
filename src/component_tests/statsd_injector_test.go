package component_tests_test

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("StatsdInjector", func() {
	var (
		consumerServer *MetronServer
		statsdAddr     string
		cleanup        func()
	)

	BeforeEach(func() {
		var err error
		consumerServer, err = NewMetronServer()
		Expect(err).ToNot(HaveOccurred())

		statsdAddr, cleanup = startStatsdInjector(fmt.Sprint(consumerServer.Port()))
	})

	AfterEach(func() {
		consumerServer.Stop() //nolint:errcheck
		cleanup()
		gexec.CleanupBuildArtifacts()
	})

	It("emits envelopes produced from statsd", func() {
		connection, err := net.Dial("udp", statsdAddr)
		Expect(err).ToNot(HaveOccurred())
		defer connection.Close()

		done := make(chan bool, 1)
		defer close(done)
		go func() {
			statsdmsg := []byte("fake-origin.test.counter:23|g")
			ticker := time.NewTicker(100 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					connection.Write(statsdmsg) //nolint:errcheck
				case <-done:
					return
				}
			}
		}()

		var receiver loggregator_v2.Ingress_SenderServer
		Eventually(consumerServer.Metron.SenderInput.Arg0).Should(Receive(&receiver))

		f := func() bool {
			e, err := receiver.Recv()
			if err != nil {
				return false
			}

			if e.GetTags()["origin"] != "fake-origin" {
				return false
			}

			return e.GetGauge().GetMetrics()["test.counter"].GetValue() == 23
		}

		Eventually(f).Should(BeTrue())
	})
})

func startStatsdInjector(metronPort string) (statsdAddr string, cleanup func()) {
	port := fmt.Sprint(testPort())

	path, err := gexec.Build("github.com/cloudfoundry/statsd-injector")
	Expect(err).ToNot(HaveOccurred())

	cmd := exec.Command(path,
		"-statsd-port", port,
		"-metron-port", metronPort,
		"-ca", CAFilePath(),
		"-cert", StatsdCertPath(),
		"-key", StatsdKeyPath(),
	)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	return fmt.Sprintf("127.0.0.1:%s", port), func() {
		session.Kill()
		session.Wait()
	}
}

func testPort() int {
	add, _ := net.ResolveTCPAddr("tcp", ":0")
	l, _ := net.ListenTCP("tcp", add)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

# statsd-injector [![slack.cloudfoundry.org][slack-badge]][loggregator-slack]
The statsd injector is a colocated job for bosh VMs that transforms metrics
from statsd format to loggregator envelope format, and sends them to the
forwarder agent on the vm. It is being maintained but not actively developed.

## Usage

The `statsd_injector` job needs to be colocated with a [Loggregator v2
envelope][loggregator-api] receiver on the
`loggregator_tls_statsdinjector.metron_port`. It receives metrics via statsd
UDP and re-emits them to the metric receiver.

Examples of loggregator v2 envelope receiver:
[loggregator forwarder agent][forwarder-agent-release]

A link to how it fits into the loggregator architecture can be found
[in the Tanzu docs](https://docs.pivotal.io/platform/application-service/2-9/loggregator/architecture.html)

### Development

The binary for `statsd_injector` is build from the code is `src/`

```bash
cd src/
go test -mod=vendor ./... -race
```

If you have ginkgo, you can use the following command:

```bash
ginkgo -r -race -randomizeAllSpecs
```

### Creating a release

This component runs as a [bosh](https://bosh.io/) job. To build a local
release:

```
bosh create-release
```

### Deployment

1. Include a certificate variable in your bosh manifest:

    ```diff
    variables:
    +  - name: loggregator_tls_statsdinjector
    +    options:
    +      ca: loggregator_ca
    +      common_name: statsdinjector
    +      extended_key_usage:
    +      - client_auth
    +    type: certificate
    ```

1. Add the release to your deployment manifest.

   ```diff
   releases:
   +  - name: statsd-injector
   +    version: latest
   ```

   Then `bosh upload release` the latest [`statsd-injector-release` bosh release][bosh-release].

1. Colocate the job in the desired instance group.

    ```diff
    instance_groups:
    - name: <targeted_instance_group>
      jobs:
    +    - name: statsd_injector
    +      release: statsd-injector
    +      properties:
    +        loggregator:
    +          tls:
    +            ca_cert: "((loggregator_tls_statsdinjector.ca))"
    +            statsd_injector:
    +              cert: "((loggregator_tls_statsdinjector.certificate))"
    +              key: "((loggregator_tls_statsdinjector.private_key))"
    ```

   Then `bosh deploy` this updated manifest.

1. Send it a metric

   You can emit statsd metrics to the injector by sending a correctly formatted
   message to udp port 8125 on the job's VM.

   As an example using `nc`:

   ```bash
   echo "origin.some.counter:1|c" | nc -u -w0 127.0.0.1 8125
   ```

   *NOTE:* The injector expects the the name of the metric to be of the form `<origin>.<metric_name>`

   The injector also supports tags according to the [Datadog StatsD extension](datadog-statsd):

   ```bash
   echo "origin.some.counter:1|c|#testtag1:testvalue1,testtag2:testvalue2" | nc -u -w0 127.0.0.1 8125
   ```

1. Validate the metric can be seen.

   Assuming you are using `statsd-injector` with CF Release, you can use the
   [CF Nozzle plugin][cf-nozzle-plugin]

   ```bash
   cf nozzle -filter CounterEvent | grep <metric_name>
   ```

   Alternatively, you could curl the metrics-agent endpoint directly:

   ```bash
   curl https://localhost:14727/metrics -k -cacert=scrape_ca.crt --cert scrape.crt --key scrape.key
   ```

[loggregator-api]:         https://github.com/cloudfoundry/loggregator-api
[grpc]:                    https://github.com/grpc/
[bosh-release]:            http://bosh.io/releases/github.com/cloudfoundry/statsd-injector-release?all=1
[datadog-statsd]:          https://docs.datadoghq.com/developers/dogstatsd/datagram_shell/
[cf-nozzle-plugin]:        https://github.com/cloudfoundry-community/firehose-plugin
[forwarder-agent-release]: https://github.com/cloudfoundry/loggregator-agent-release

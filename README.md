# statsd-injector
Companion component to Metron that receives Statsd and emits Dropsonde to Metron

## Including statsd-injector in a bosh deployment
As an example, if you want the injector to be present on loggregator boxes, add the following in `cf-lamb.yml`

```diff
   instance_groups:
   - name: doppler
     jobs:
     - name: doppler
       release: loggregator
     - name: metron_agent
       release: loggregator
+    - name: statsd-injector
+      release: statsd-injector
```

## Emitting metrics to the statsd-injector
You can emit statsd metrics to the injector by sending a correctly formatted message to udp port 8125

As an example using netcat:

```
echo "origin.some.counter:1|c" | nc -u -w0 127.0.0.1 8125
```

You should see the metric come out of the firehose.

The injector expects the the name of the metric to be of the form `<origin>.<metric_name>`

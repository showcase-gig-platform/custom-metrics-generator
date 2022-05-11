# custom-metrics-generator

`custom-metrics-generator` is generate prometheus metrics using custom resource definition.  
You can output any metrics to http endpoints like a prometheus exporter.

## Image

### Official

`public.ecr.aws/q1m5p9s1/custom-metrics-generator` (amd64 linux only)

### Build

`$ IMG=<<your repository>>:<<tag>> make docker-build`

## Deploy CRD and controller

See `manifest/deploy`.  

`$ kustomize build manifests/deploy | kubectl apply -f -`

### Flags

```
-generate-metrics-bind-address string
    Generated metrics endpoint addr. (default ":8082")
-generate-metrics-path string
    Generated metrics path. (default "/metrics")
-interval-seconds int
    interval seconds to fetch metrics (default 60)
-metrics-prefix string
    set prefix for metrics name (default none)
-offset-seconds int
    offset seconds to generate metrics (default 0)
-timezone string
    set timezone (default "UTC")
```

## Deploy resources

Sample is in `manifest/resource`.  

### Fields

Name|Type|Required|Description
---|---|---|---
spec.metricsName|string|Yes|Name of generated metrics.
spec.offsetSeconds|int|No|Offset seconds to generate metrics (override flag setting)
spec.timezone|string|No|Set timezone (override flag setting)
spec.labels|map[string]string|No|Labels to be added to generated metrics.
spec.metrics.start|string|Yes| __Cron formatted__ schedule to start output metrics.
spec.metrics.duration|duration|Yes|Duration to keep output metrics.
spec.metrics.value|int|Yes|Value of output metrics.

### Rules of define metrics

#### Naming

`spec.metricsName` must match the regex `[a-zA-Z_:][a-zA-Z0-9_:]*`.  
Keys in `spec.labels` must match the regex `[a-zA-Z_][a-zA-Z0-9_]*` and cannot start with two or more `_`.  
See also [prometheus docs](https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels).

If invalid character is used, it is replaced by `_`.

### Multiple metrics

`spec.metrics` field is specified as array, so you can define more than one.  
If multiple metrics have overlapping schedules, new schedule metrics will be used.

![metrics sample](images/sample.png)

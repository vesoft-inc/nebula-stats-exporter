module main

require (
	github.com/nebula-stats-exporter/exporter v0.0.1
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.0
	k8s.io/client-go v0.17.0
)

replace (
	github.com/nebula-stats-exporter/exporter => ../nebula-stats-exporter/exporter
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/metrics => k8s.io/metrics v0.17.0
)

go 1.13

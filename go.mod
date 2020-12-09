module github.com/vesoft-inc/nebula-stats-exporter

go 1.13

require (
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog v1.0.0

)

replace (
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/metrics => k8s.io/metrics v0.18.6
)

package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type metrics struct {
	MetricsType string
	desc        *prometheus.Desc
}

type NebulaExporter struct {
	sync.Mutex
	Client *kubernetes.Clientset
	mux    *http.ServeMux

	Namespace               string
	ListenAddress           string
	exporterMetricsRegistry *prometheus.Registry
	config                  StaticConfig
	graphdMap               map[string]metrics
	storagedMap             map[string]metrics
	metadMap                map[string]metrics
}

func (exporter *NebulaExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	exporter.mux.ServeHTTP(w, r)
}

func getNebulaMetrics(ipAddress string, port int32) ([]string, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/get_stats?stats", ipAddress, port))

	if err != nil {
		return []string{}, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return []string{}, err
	}

	metricStr := string(bytes)

	metrics := strings.Split(metricStr, "\n")

	return metrics, nil
}

func NewNebulaExporter(ns string, listenAddr string, client *kubernetes.Clientset,
	config StaticConfig, maxRequests int) (*NebulaExporter, error) {
	exporter := &NebulaExporter{
		Namespace:               ns,
		ListenAddress:           listenAddr,
		Client:                  client,
		config:                  config,
		graphdMap:               make(map[string]metrics),
		storagedMap:             make(map[string]metrics),
		metadMap:                make(map[string]metrics),
		exporterMetricsRegistry: prometheus.NewRegistry(),
	}

	secondaryLabels := []string{
		"latency", "error_qps", "all_qps", "qps",
	}

	thirdLabels := []string{
		"count", "avg", "rate", "sum",
	}

	lastLabels := []string{
		"5", "60", "600", "3600",
	}

	graphdLabels := []string{
		"create_edge", "fetch_vertices", "parse",
		"create_tag", "all", "create_space",
		"drop_space", "go", "insert_edge",
		"storageClient", "use", "insert_vertex",
		"show", "yield",
	}

	for _, graphdLabel := range graphdLabels {
		for _, secondaryLabel := range secondaryLabels {
			for _, thirdLabel := range thirdLabels {
				for _, lastLabel := range lastLabels {
					exporter.graphdMap["graph_"+graphdLabel+"_"+secondaryLabel+"."+thirdLabel+"."+lastLabel] = metrics{
						MetricsType: thirdLabel,
						desc: prometheus.NewDesc(
							prometheus.BuildFQName("nebula", "graphd", graphdLabel+"_"+secondaryLabel+"_"+thirdLabel+"_"+lastLabel),
							"", []string{"instanceName", "namespace", "instanceType"},
							nil),
					}
				}
			}
		}
	}

	exporter.graphdMap["graphd_count"] = metrics{
		MetricsType: "graphd_count",
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nebula", "graphd", "count"),
			"", []string{"instanceName", "namespace", "instanceType"}, nil),
	}


	metadLabels := []string{
		"heartbeat",
	}

	for _, metadLabel := range metadLabels {
		for _, secondaryLabel := range secondaryLabels {
			for _, thirdLabel := range thirdLabels {
				for _, lastLabel := range lastLabels {
					exporter.metadMap["meta_"+metadLabel+"_"+secondaryLabel+"."+thirdLabel+"."+lastLabel] = metrics{
						MetricsType: thirdLabel,
						desc: prometheus.NewDesc(
							prometheus.BuildFQName("nebula", "metad", metadLabel+"_"+secondaryLabel+"_"+thirdLabel+"_"+lastLabel),
							"", []string{"instanceName", "namespace", "instanceType"}, nil),
					}
				}
			}
		}
	}

	exporter.metadMap["metad_count"] = metrics{
		MetricsType: "metad_count",
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nebula", "metad", "count"),
			"", []string{"instanceName", "namespace", "instanceType"}, nil),
	}

	storagedLabels := []string{
		"get_bound", "bound_stats", "vertex_props",
		"edge_props", "add_vertex", "add_edge",
		"drop_space", "go", "insert_edge",
		"del_vertex", "update_vertex", "update_edge",
		"get_kv", "put_kv",
	}

	for _, storagedLabel := range storagedLabels {
		for _, secondaryLabel := range secondaryLabels {
			for _, thirdLabel := range thirdLabels {
				for _, lastLabel := range lastLabels {
					exporter.storagedMap["storage_"+storagedLabel+"_"+secondaryLabel+"."+thirdLabel+"."+lastLabel] = metrics{
						MetricsType: thirdLabel,
						desc: prometheus.NewDesc(
							prometheus.BuildFQName("nebula", "storage", storagedLabel+"_"+secondaryLabel+"_"+thirdLabel+"_"+lastLabel),
							"", []string{"instanceName", "namespace", "instanceType"}, nil),
					}

				}
			}
		}
	}

	exporter.storagedMap["storaged_count"] = metrics{
		MetricsType: "storage_count",
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nebula", "storaged", "count"),
			"", []string{"instanceName", "namespace", "instanceType"}, nil),
	}

	err := exporter.exporterMetricsRegistry.Register(exporter)

	if err != nil {
		log.Fatalf("Register Nebula Collector Failed: %v", err)
		return nil, err
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{exporter.exporterMetricsRegistry},
		promhttp.HandlerOpts{
			ErrorLog:            log.NewErrorLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
			Registry:            exporter.exporterMetricsRegistry,
		},
	)

	metricsHandler := promhttp.InstrumentMetricHandler(
		exporter.exporterMetricsRegistry, handler,
	)

	exporter.mux = http.NewServeMux()

	exporter.mux.Handle("/metrics", metricsHandler)

	exporter.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	exporter.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Nebula Exporter </title></head>
			<body>
				<h1>Nebula Exporter </h1>
				<p><a href='metrics'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	exporter.exporterMetricsRegistry = prometheus.NewRegistry()

	return exporter, nil
}
func (exporter *NebulaExporter) Describe(ch chan<- *prometheus.Desc) {
	log.Info("Begin Describe Nebula Metrics")

	for _, metrics := range exporter.storagedMap {
		ch <- metrics.desc
	}

	for _, metrics := range exporter.metadMap {
		ch <- metrics.desc
	}

	for _, metrics := range exporter.graphdMap {
		ch <- metrics.desc
	}
	log.Info("Describe Nebula Metrics Done")
}

func (exporter *NebulaExporter) Collect(ch chan<- prometheus.Metric) {
	if exporter.Client != nil {
		exporter.CollectFromKuberneres(ch)
	} else {
		exporter.CollectFromStaticConfig(ch)
	}
}

func (exporter *NebulaExporter) CollectMetrics(name, typ, namespace string, metrics []string, ch chan<- prometheus.Metric) {
	if typ == "metad" {
		log.Infof("Begin Collect Metad %s Metrics ", name)
		for _, singleMetric := range metrics {
			log.Info(singleMetric)
			seps := strings.Split(singleMetric, "=")
			if len(seps) != 2 {
				continue
			}

			values, err := strconv.ParseFloat(seps[1], 64)
			if err != nil {
				continue
			}

			if metric, ok := exporter.metadMap[seps[0]]; ok {
				ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, values,
					name,
					namespace,
					typ,
				)
			}
		}

	} else if typ == "storaged" {
		log.Infof("Begin Collect Storaged %s Metrics ", name)
		for _, singleMetric := range metrics {
			log.Info(singleMetric)
			seps := strings.Split(singleMetric, "=")
			if len(seps) != 2 {
				continue
			}

			values, err := strconv.ParseFloat(seps[1], 64)
			if err != nil {
				continue
			}

			if metric, ok := exporter.storagedMap[seps[0]]; ok {
				ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, values,
					name,
					namespace,
					typ,
				)
			}
		}
	} else if typ == "graphd" {
		log.Infof("Begin Collect Graphd %s Metrics ", name)
		for _, singleMetric := range metrics {
			log.Info(singleMetric)
			seps := strings.Split(singleMetric, "=")
			if len(seps) != 2 {
				continue
			}

			values, err := strconv.ParseFloat(seps[1], 64)
			if err != nil {
				continue
			}

			if metric, ok := exporter.graphdMap[seps[0]]; ok {
				ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, values,
					name,
					namespace,
					typ,
				)
			}
		}
	}
}

func (exporter *NebulaExporter) CollectFromStaticConfig(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	for _, item := range exporter.config.NebulaItems {
		podIpAddress := item.EndpointIP
		podHttpPort := item.EndpointPort

		if podHttpPort == 0 {
			continue
		}

		wg.Add(1)
		go func() {
			metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)

			if err != nil {
				wg.Done()
				return
			}
			exporter.CollectMetrics(item.InstanceName, item.NebulaType, "nebula", metrics, ch)

			wg.Done()
		}()

		wg.Wait()
		log.Info("Collect Nebula Metrics Done")
	}
}

func (exporter *NebulaExporter) CollectFromKuberneres(ch chan<- prometheus.Metric) {
	log.Info("Begin Collect Nebula Metrics")
	podLists, err := exporter.Client.CoreV1().Pods(exporter.Namespace).List(metav1.ListOptions{})

	if err != nil {
		return
	}

	var wg sync.WaitGroup

	for _, pod := range podLists.Items {
		podIpAddress := pod.Status.PodIP
		podHttpPort := int32(0)
		for _, port := range pod.Spec.Containers[0].Ports {
			if port.Name == "http" {
				podHttpPort = port.ContainerPort
			}
		}

		if podHttpPort == 0 {
			continue
		}

		wg.Add(1)

		go func() {
			metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)

			if err != nil {
				wg.Done()
				return
			}

			nebulaType := ""
			if strings.Contains(pod.Name, "metad") {
				nebulaType = "metad"
			} else if strings.Contains(pod.Name, "storaged") {
				nebulaType = "storaged"
			} else if strings.Contains(pod.Name, "graphd") {
				nebulaType = "graphd"
			}

			exporter.CollectMetrics(pod.Name, nebulaType, pod.Namespace, metrics, ch)

			wg.Done()
		}()

		wg.Wait()
		log.Info("Collect Nebula Metrics Done")
	}
}

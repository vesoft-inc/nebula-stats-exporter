package exporter

import (
	"context"
	"encoding/json"
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
	"k8s.io/klog"
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

func getNebulaComponentStatus(ipAddress string, port int32, nebulaType string) ([]string, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/status", ipAddress, port))

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	type nebulaStatus struct {
		GitInfoSha string `json:"git_info_sha"`
		Status     string `json:"status"`
	}

	var status nebulaStatus
	if err := json.Unmarshal(bytes, &status); err != nil {
		return nil, err
	}

	countStr := fmt.Sprintf("%s_count=0", nebulaType)
	if status.Status == "running" {
		countStr = strings.Replace(countStr, "0", "1", 1)
	}
	statusMetrics := []string{countStr}

	return statusMetrics, nil
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
		"latency", "error_qps", "qps",
	}

	thirdLabels := []string{
		"count", "avg", "rate", "sum", "p99",
	}

	lastLabels := []string{
		"5", "60", "600", "3600",
	}

	graphdLabels := []string{
		"storageClient",
		"metaClient",
		"all",
		"use",
		"show",
		"find_path",
		"create_edge",
		"create_tag",
		"create_snapshot",
		"delete_edge",
		"delete_vertices",
		"update_vertex",
		"update_edge",
		"alter_tag",
		"alter_edge",
		"drop_tag",
		"drop_edge",
		"drop_space",
		"drop_snapshot",
		"insert_edge",
		"insert_vertex",
		"fetch_edges",
		"fetch_vertices",
		"describe_tag",
		"describe_edge",
		"describe_space",
		"go",
		"return",
		"set",
		"yield",
		"group_by",
		"order_by",
		"limit",
		"assignment",
		"ingest",
		"config",
		"balance",
		"download",
	}

	for _, graphdLabel := range graphdLabels {
		for _, secondaryLabel := range secondaryLabels {
			for _, thirdLabel := range thirdLabels {
				if thirdLabel == "p99" {
					if secondaryLabel == "error_qps" || secondaryLabel == "qps" {
						continue
					}
				}
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
		MetricsType: "count",
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
		MetricsType: "count",
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nebula", "metad", "count"),
			"", []string{"instanceName", "namespace", "instanceType"}, nil),
	}

	storagedLabels := []string{
		"get_bound",
		"bound_stats",
		"vertex_props",
		"edge_props",
		"add_vertex",
		"add_edge",
		"del_vertex",
		"update_vertex",
		"update_edge",
		"get_kv",
		"put_kv",
		"lookup_edges",
		"lookup_vertices",
		"scan_vertex",
		"scan_edge",
	}

	for _, storagedLabel := range storagedLabels {
		for _, secondaryLabel := range secondaryLabels {
			for _, thirdLabel := range thirdLabels {
				if thirdLabel == "p99" {
					if secondaryLabel == "error_qps" || secondaryLabel == "qps" {
						continue
					}
				}
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
		MetricsType: "count",
		desc: prometheus.NewDesc(
			prometheus.BuildFQName("nebula", "storaged", "count"),
			"", []string{"instanceName", "namespace", "instanceType"}, nil),
	}

	err := exporter.exporterMetricsRegistry.Register(exporter)

	if err != nil {
		klog.Fatalf("Register Nebula Collector Failed: %v", err)
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
	klog.Info("Begin Describe Nebula Metrics")

	for _, metrics := range exporter.storagedMap {
		ch <- metrics.desc
	}

	for _, metrics := range exporter.metadMap {
		ch <- metrics.desc
	}

	for _, metrics := range exporter.graphdMap {
		ch <- metrics.desc
	}
	klog.Info("Describe Nebula Metrics Done")
}

func (exporter *NebulaExporter) Collect(ch chan<- prometheus.Metric) {
	if exporter.Client != nil {
		exporter.CollectFromKubernetes(ch)
	} else {
		exporter.CollectFromStaticConfig(ch)
	}
}

func (exporter *NebulaExporter) CollectMetrics(
	name string,
	nebulaType string,
	namespace string,
	metrics []string,
	ch chan<- prometheus.Metric) {
	if nebulaType == "metad" {
		for _, singleMetric := range metrics {
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
					nebulaType,
				)
			}
		}

	} else if nebulaType == "storaged" {
		for _, singleMetric := range metrics {
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
					nebulaType,
				)
			}
		}
	} else if nebulaType == "graphd" {
		for _, singleMetric := range metrics {
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
					nebulaType,
				)
			}
		}
	}
}

func (exporter *NebulaExporter) CollectFromStaticConfig(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(len(exporter.config.NebulaItems))

	for _, nebulaItem := range exporter.config.NebulaItems {
		item := nebulaItem
		podIpAddress := item.EndpointIP
		podHttpPort := item.EndpointPort

		if podHttpPort == 0 {
			continue
		}

		go func() {
			defer wg.Done()
			klog.Infof("Collect %s %s Metrics ", strings.ToUpper(item.NebulaType), item.InstanceName)
			metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
			if len(metrics) == 0 {
				klog.Infof("metrics from %s:%d was empty", podIpAddress, podHttpPort)
			}
			if err != nil {
				klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
				return
			}
			exporter.CollectMetrics(item.InstanceName, item.NebulaType, "nebula", metrics, ch)
		}()

		statusMetrics, err := getNebulaComponentStatus(podIpAddress, podHttpPort, item.NebulaType)
		if err != nil {
			klog.Errorf("get status metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
			return
		}
		exporter.CollectMetrics(item.InstanceName, item.NebulaType, "nebula", statusMetrics, ch)
	}

	wg.Wait()
	klog.Info("Collect Nebula Metrics Done")
}

func (exporter *NebulaExporter) CollectFromKubernetes(ch chan<- prometheus.Metric) {
	klog.Info("Begin Collect Nebula Metrics")

	podLists, err := exporter.Client.CoreV1().Pods(exporter.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(podLists.Items))

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

		nebulaType := ""
		if strings.Contains(pod.Name, "metad") {
			nebulaType = "metad"
		} else if strings.Contains(pod.Name, "storaged") {
			nebulaType = "storaged"
		} else if strings.Contains(pod.Name, "graphd") {
			nebulaType = "graphd"
		}

		go func() {
			defer wg.Done()
			klog.Infof("Collect %s %s Metrics ", strings.ToUpper(nebulaType), pod.Name)
			metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
			if len(metrics) == 0 {
				klog.Infof("metrics from %s:%d was empty", podIpAddress, podHttpPort)
			}
			if err != nil {
				klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
				return
			}

			exporter.CollectMetrics(pod.Name, nebulaType, pod.Namespace, metrics, ch)
		}()

		statusMetrics, err := getNebulaComponentStatus(podIpAddress, podHttpPort, nebulaType)
		if len(statusMetrics) == 0 {
			klog.Errorf("could not get status metrics from %s:%d", podIpAddress, podHttpPort)
		}
		if err != nil {
			klog.Errorf("get status metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
			return
		}
		exporter.CollectMetrics(pod.Name, nebulaType, pod.Namespace, statusMetrics, ch)
	}

	wg.Wait()
	klog.Info("Collect Nebula Metrics Done")
}

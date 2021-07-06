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

type ComponentType string

const (
	ComponentLabelKey = "app.kubernetes.io/component"
	ClusterLabelKey   = "app.kubernetes.io/cluster"

	GraphdComponent   ComponentType = "graphd"
	MetadComponent    ComponentType = "metad"
	StoragedComponent ComponentType = "storaged"

	// FQNamespace represents the prometheus FQName
	FQNamespace            = "nebula"
	DefaultStaticNamespace = "nebula"
)

func (ct ComponentType) String() string {
	return string(ct)
}

type metrics struct {
	metricsType string
	desc        *prometheus.Desc
}

type NebulaExporter struct {
	client        *kubernetes.Clientset
	mux           *http.ServeMux
	namespace     string
	cluster       string
	listenAddress string
	registry      *prometheus.Registry
	config        StaticConfig
	metricMap     map[string]metrics
}

func (exporter *NebulaExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	exporter.mux.ServeHTTP(w, r)
}

func getNebulaMetrics(ipAddress string, port int32) ([]string, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/stats?stats=", ipAddress, port))
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	metricStr := string(bytes)
	metrics := strings.Split(metricStr, "\n")
	if len(metrics) > 0 {
		if metrics[len(metrics)-1] == "" {
			metrics = metrics[:len(metrics)-1]
		}
	}

	return metrics, nil
}

func getNebulaComponentStatus(ipAddress string, port int32, componentType string) ([]string, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s:%d/status", ipAddress, port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	countStr := fmt.Sprintf("%s_count=0", componentType)
	if status.Status == "running" {
		countStr = strings.Replace(countStr, "0", "1", 1)
	}
	statusMetrics := []string{countStr}

	return statusMetrics, nil
}

func NewNebulaExporter(ns, cluster, listenAddr string, client *kubernetes.Clientset,
	config StaticConfig, maxRequests int) (*NebulaExporter, error) {
	exporter := &NebulaExporter{
		namespace:     ns,
		cluster:       cluster,
		listenAddress: listenAddr,
		client:        client,
		config:        config,
		metricMap:     make(map[string]metrics),
		registry:      prometheus.NewRegistry(),
	}

	secondaryLabels := []string{
		"latency_us",
	}

	thirdLabels := []string{
		"rate", "sum", "avg", "p75", "p95", "p99", "p999",
	}

	lastLabels := []string{
		"5", "60", "600", "3600",
	}

	graphdLabels := []string{
		"slow_query",
		"query",
		"num_query_errors",
		"num_queries",
		"num_slow_queries",
	}

	for _, graphdLabel := range graphdLabels {
		if strings.HasPrefix(graphdLabel, "num_") {
			for _, secondLabel := range thirdLabels[:2] {
				for _, lastLabel := range lastLabels {
					exporter.buildMetricMap(GraphdComponent, graphdLabel, secondLabel, lastLabel)
				}
			}
		} else {
			for _, secondLabel := range secondaryLabels {
				for _, thirdLabel := range thirdLabels[2:] {
					for _, lastLabel := range lastLabels {
						exporter.buildMetricMap(GraphdComponent, graphdLabel, secondLabel, thirdLabel, lastLabel)
					}
				}
			}
		}
	}

	metadLabels := []string{
		"heartbeat",
		"num_heartbeats",
	}

	for _, metadLabel := range metadLabels {
		if strings.HasPrefix(metadLabel, "num_") {
			for _, secondLabel := range thirdLabels[:2] {
				for _, lastLabel := range lastLabels {
					exporter.buildMetricMap(MetadComponent, metadLabel, secondLabel, lastLabel)
				}
			}
		} else {
			for _, secondLabel := range secondaryLabels {
				for _, thirdLabel := range thirdLabels[2 : len(thirdLabels)-1] {
					for _, lastLabel := range lastLabels {
						exporter.buildMetricMap(MetadComponent, metadLabel, secondLabel, thirdLabel, lastLabel)
					}
				}
			}
		}
	}

	storagedLabels := []string{
		"num_lookup_errors",
		"num_lookup",
		"num_get_prop_errors",
		"num_get_neighbors",
		"update_edge",
		"num_update_edge_errors",
		"num_scan_edge",
		"num_update_edge",
		"scan_edge",
		"update_vertex",
		"num_scan_vertex_errors",
		"get_value",
		"num_add_vertices_errors",
		"add_edges_atomic",
		"num_forward_tranx_errors",
		"num_delete_vertices_errors",
		"num_scan_edge_errors",
		"num_delete_edges",
		"scan_vertex",
		"num_add_edges",
		"num_delete_vertices",
		"num_add_edges_atomic",
		"num_add_vertices",
		"num_add_edges_atomic_errors",
		"num_get_value",
		"num_get_neighbors_errors",
		"add_vertices",
		"num_get_prop",
		"forward_tranx",
		"num_get_value_errors",
		"add_edges",
		"lookup",
		"delete_edges",
		"num_add_edges_errors",
		"num_update_vertex_errors",
		"num_scan_vertex",
		"get_prop",
		"get_neighbors",
		"delete_vertices",
		"num_forward_tranx",
		"num_delete_edges_errors",
		"num_update_vertex",
	}

	for _, storagedLabel := range storagedLabels {
		if strings.HasPrefix(storagedLabel, "num_") {
			for _, secondLabel := range thirdLabels[:2] {
				for _, lastLabel := range lastLabels {
					exporter.buildMetricMap(StoragedComponent, storagedLabel, secondLabel, lastLabel)
				}
			}
		} else {
			for _, secondLabel := range secondaryLabels {
				for _, thirdLabel := range thirdLabels[2 : len(thirdLabels)-1] {
					for _, lastLabel := range lastLabels {
						exporter.buildMetricMap(StoragedComponent, storagedLabel, secondLabel, thirdLabel, lastLabel)
					}
				}
			}
		}
	}

	t := []string{GraphdComponent.String(), MetadComponent.String(), StoragedComponent.String()}
	for _, c := range t {
		k := fmt.Sprintf("%s_count", c)
		exporter.metricMap[k] = metrics{
			metricsType: "count",
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(FQNamespace, c, "count"),
				"", []string{"instanceName", "namespace", "cluster", "componentType"}, nil),
		}
	}

	if err := exporter.registry.Register(exporter); err != nil {
		klog.Fatalf("Register Nebula Collector Failed: %v", err)
		return nil, err
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{exporter.registry},
		promhttp.HandlerOpts{
			ErrorLog:            log.NewErrorLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
			Registry:            exporter.registry,
		},
	)

	metricsHandler := promhttp.InstrumentMetricHandler(
		exporter.registry, handler,
	)

	exporter.mux = http.NewServeMux()

	exporter.mux.Handle("/metrics", metricsHandler)

	exporter.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	exporter.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
			<html>
			<head><title>Nebula Exporter </title></head>
			<body>
				<h1>Nebula Exporter </h1>
				<p><a href='metrics'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	exporter.registry = prometheus.NewRegistry()

	return exporter, nil
}

func (exporter *NebulaExporter) Describe(ch chan<- *prometheus.Desc) {
	klog.Info("Begin Describe Nebula Metrics")

	for _, metrics := range exporter.metricMap {
		ch <- metrics.desc
	}

	klog.Info("Describe Nebula Metrics Done")
}

func (exporter *NebulaExporter) Collect(ch chan<- prometheus.Metric) {
	klog.Infoln("Collect!")
	if exporter.client != nil {
		exporter.CollectFromKubernetes(ch)
	} else {
		exporter.CollectFromStaticConfig(ch)
	}
}

func (exporter *NebulaExporter) CollectMetrics(
	name string,
	componentType string,
	namespace string,
	cluster string,
	metrics []string,
	ch chan<- prometheus.Metric) {
	if len(metrics) == 0 {
		return
	}

	// status metric
	if len(metrics) == 1 {
		seps := strings.Split(metrics[0], "=")
		values, err := strconv.ParseFloat(seps[1], 64)
		if err == nil {
			if metric, ok := exporter.metricMap[seps[0]]; ok {
				ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, values,
					name,
					namespace,
					cluster,
					componentType,
				)
			}
		}
		return
	}

	// query metrics
	matches := convertToMap(metrics)
	for metric, value := range matches {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		if metric, ok := exporter.metricMap[metric]; ok {
			ch <- prometheus.MustNewConstMetric(metric.desc, prometheus.GaugeValue, v,
				name,
				namespace,
				cluster,
				componentType,
			)
		}
	}
}

func (exporter *NebulaExporter) CollectFromStaticConfig(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, cluster := range exporter.config.Clusters {
		cluster := cluster
		wg.Add(len(cluster.Instances))
		for _, instance := range cluster.Instances {
			instance := instance
			podIpAddress := instance.EndpointIP
			podHttpPort := instance.EndpointPort

			if podHttpPort == 0 {
				wg.Done()
				continue
			}

			go func() {
				defer wg.Done()
				klog.Infof("Collect %s:%s %s:%d Metrics ",
					cluster.Name, strings.ToUpper(instance.ComponentType),
					instance.Name, instance.EndpointPort)
				metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
				if len(metrics) == 0 {
					klog.Infof("metrics from %s:%d was empty", podIpAddress, podHttpPort)
				}
				if err != nil {
					klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
					return
				}
				exporter.CollectMetrics(instance.Name, instance.ComponentType, DefaultStaticNamespace, cluster.Name, metrics, ch)
			}()

			statusMetrics, err := getNebulaComponentStatus(podIpAddress, podHttpPort, instance.ComponentType)
			if err != nil {
				klog.Errorf("get status metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
				return
			}
			exporter.CollectMetrics(instance.Name, instance.ComponentType, DefaultStaticNamespace, cluster.Name, statusMetrics, ch)
		}
	}

	wg.Wait()
	klog.Info("Collect Nebula Metrics Done")
}

func (exporter *NebulaExporter) CollectFromKubernetes(ch chan<- prometheus.Metric) {
	klog.Info("Begin Collect Nebula Metrics")

	listOpts := metav1.ListOptions{}
	if exporter.cluster != "" {
		listOpts.LabelSelector = fmt.Sprintf("%s=%s", ClusterLabelKey, exporter.cluster)
	}
	podLists, err := exporter.client.CoreV1().Pods(exporter.namespace).List(context.TODO(), listOpts)
	if err != nil {
		klog.Error(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(podLists.Items))

	for _, item := range podLists.Items {
		pod := item

		componentType, ok := pod.Labels[ComponentLabelKey]
		if !ok {
			wg.Done()
			continue
		}

		cluster, ok := pod.Labels[ClusterLabelKey]
		if !ok {
			wg.Done()
			continue
		}

		podIpAddress := pod.Status.PodIP
		podHttpPort := int32(0)
		for _, port := range pod.Spec.Containers[0].Ports {
			if port.Name == "http" {
				podHttpPort = port.ContainerPort
			}
		}

		if podHttpPort == 0 {
			wg.Done()
			continue
		}

		go func() {
			defer wg.Done()
			klog.Infof("Collect %s %s:%d Metrics ", strings.ToUpper(componentType), pod.Name, podHttpPort)
			metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
			if len(metrics) == 0 {
				klog.Infof("metrics from %s:%d was empty", podIpAddress, podHttpPort)
			}
			if err != nil {
				klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
				return
			}

			exporter.CollectMetrics(pod.Name, componentType, pod.Namespace, cluster, metrics, ch)
		}()

		statusMetrics, err := getNebulaComponentStatus(podIpAddress, podHttpPort, componentType)
		if len(statusMetrics) == 0 {
			klog.Errorf("could not get status metrics from %s:%d", podIpAddress, podHttpPort)
		}
		if err != nil {
			klog.Errorf("get status metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
			return
		}
		exporter.CollectMetrics(pod.Name, componentType, pod.Namespace, cluster, statusMetrics, ch)
	}

	wg.Wait()
	klog.Info("Collect Nebula Metrics Done")
}

func (exporter *NebulaExporter) buildMetricMap(
	component ComponentType,
	labels ...string) *NebulaExporter {
	if len(labels) == 0 {
		return exporter
	}

	var k string
	var metricName string

	last := len(labels) - 2
	if last <= 0 {
		return exporter
	}

	for i, label := range labels {
		if i == 0 {
			metricName = label
			k = label
		} else {
			metricName = metricName + "_" + label
			if i < last {
				k = k + "_" + label
			}
		}
	}

	for _, label := range labels[last:] {
		k = k + "." + label
	}

	exporter.metricMap[k] = metrics{
		metricsType: labels[last],
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(FQNamespace, component.String(), metricName),
			"", []string{"instanceName", "namespace", "cluster", "componentType"},
			nil),
	}
	return exporter
}

func convertToMap(metrics []string) map[string]string {
	matches := make(map[string]string)
	for t := 0; ; t += 2 {
		if t+1 >= len(metrics) {
			break
		}
		ok, value := getRValue(metrics[t])
		if !ok {
			continue
		}
		ok, metric := getRValue(metrics[t+1])
		if !ok {
			continue
		}
		if value != "" && metric != "" {
			matches[metric] = value
		}
	}
	return matches
}

func getRValue(metric string) (bool, string) {
	seps := strings.Split(metric, "=")
	if len(seps) != 2 {
		return false, ""
	}
	return true, seps[1]
}

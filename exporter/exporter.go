package exporter

import (
	"context"
	"fmt"
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

type NebulaExporter struct {
	client        *kubernetes.Clientset
	mux           *http.ServeMux
	namespace     string
	cluster       string
	listenAddress string
	registry      *prometheus.Registry
	config        StaticConfig
}

func NewNebulaExporter(ns, cluster, listenAddr string, client *kubernetes.Clientset,
	config StaticConfig, maxRequests int) (*NebulaExporter, error) {
	exporter := &NebulaExporter{
		namespace:     ns,
		cluster:       cluster,
		listenAddress: listenAddr,
		client:        client,
		config:        config,
		mux:           http.NewServeMux(),
		registry:      prometheus.NewRegistry(),
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

	return exporter, nil
}

func (exporter *NebulaExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	exporter.mux.ServeHTTP(w, r)
}

func (exporter *NebulaExporter) Describe(ch chan<- *prometheus.Desc) {}

func (exporter *NebulaExporter) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	klog.Infoln("Start collect")
	defer func() {
		klog.Infof("Complete collect, time elapse %s", time.Since(now))
	}()

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
	originMetrics []string,
	ch chan<- prometheus.Metric) {
	if len(originMetrics) == 0 {
		return
	}

	metrics := convertToMetrics(originMetrics)
	for _, metric := range metrics {
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			continue
		}

		labels := []string{"nebula_cluster", "component_type", "instance_name"}
		labelValues := []string{cluster, componentType, name}

		if namespace != NonNamespace {
			labels = append(labels, "namespace")
			labelValues = append(labelValues, namespace)
		}

		for key, value := range metric.Labels {
			labels = append(labels, key)
			labelValues = append(labelValues, value)
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				fmt.Sprintf("%s_%s_%s", FQNamespace, componentType, strings.ReplaceAll(metric.Name, ".", "_")),
				"",
				labels,
				nil,
			),
			prometheus.GaugeValue,
			v,
			labelValues...,
		)
	}
}

func (exporter *NebulaExporter) collect(wg *sync.WaitGroup, namespace, clusterName string, instance Instance, ch chan<- prometheus.Metric) {
	podIpAddress := instance.EndpointIP
	podHttpPort := instance.EndpointPort

	if podHttpPort <= 0 {
		return
	}

	klog.Infof("Collect %s:%s %s:%d Metrics ",
		clusterName, strings.ToUpper(instance.ComponentType),
		instance.Name, instance.EndpointPort)

	wg.Add(2)
	go func() {
		defer wg.Done()
		metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
		if err != nil {
			klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
			return
		}
		exporter.CollectMetrics(instance.Name, instance.ComponentType, namespace, clusterName, metrics, ch)
	}()

	go func() {
		defer wg.Done()
		statusMetrics := "count=1"
		if !isNebulaComponentRunning(podIpAddress, podHttpPort) {
			statusMetrics = "count=0"
		}
		exporter.CollectMetrics(instance.Name, instance.ComponentType, namespace, clusterName, []string{statusMetrics}, ch)
	}()
}

func (exporter *NebulaExporter) CollectFromStaticConfig(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, cluster := range exporter.config.Clusters {
		cluster := cluster
		if cluster.Name == "" {
			cluster.Name = DefaultClusterName
		}
		for _, instance := range cluster.Instances {
			instance := instance
			if instance.Name == "" {
				instance.Name = fmt.Sprintf("%s-%s", instance.EndpointIP, instance.ComponentType)
			}
			exporter.collect(&wg, NonNamespace, cluster.Name, instance, ch)
		}
	}

	wg.Wait()
}

func (exporter *NebulaExporter) CollectFromKubernetes(ch chan<- prometheus.Metric) {
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
	for _, item := range podLists.Items {
		pod := item

		componentType, ok := pod.Labels[ComponentLabelKey]
		if !ok {
			continue
		}

		clusterName, ok := pod.Labels[ClusterLabelKey]
		if !ok {
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
			continue
		}

		exporter.collect(&wg, pod.Namespace, clusterName, Instance{
			Name:          pod.Name,
			EndpointIP:    podIpAddress,
			EndpointPort:  podHttpPort,
			ComponentType: componentType,
		}, ch)
	}

	wg.Wait()
}

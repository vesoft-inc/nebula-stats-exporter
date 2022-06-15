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
	Client          *kubernetes.Clientset
	Namespace       string
	Selector        string
	Cluster         string
	ClusterLabelKey string
	GraphPortName   string
	MetaPortName    string
	StoragePortName string
	ListenAddress   string
	Config          StaticConfig
	registry        *prometheus.Registry
	mux             *http.ServeMux
}

func (e *NebulaExporter) Initialize(maxRequests int) error {
	if e.ClusterLabelKey == "" {
		e.ClusterLabelKey = ClusterLabelKey
	}

	e.mux = http.NewServeMux()
	e.registry = prometheus.NewRegistry()

	if err := e.registry.Register(e); err != nil {
		klog.Fatalf("Register Nebula Collector Failed: %v", err)
		return err
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{e.registry},
		promhttp.HandlerOpts{
			ErrorLog:            log.NewErrorLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: maxRequests,
			Registry:            e.registry,
		},
	)

	metricsHandler := promhttp.InstrumentMetricHandler(
		e.registry, handler,
	)

	e.mux.Handle("/metrics", metricsHandler)

	e.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	e.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	return nil
}

func (e *NebulaExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mux.ServeHTTP(w, r)
}

func (e *NebulaExporter) Describe(ch chan<- *prometheus.Desc) {}

func (e *NebulaExporter) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	klog.Infoln("Start collect")
	defer func() {
		klog.Infof("Complete collect, time elapse %s", time.Since(now))
	}()

	if e.Client != nil {
		e.CollectFromKubernetes(ch)
	} else {
		e.CollectFromStaticConfig(ch)
	}
}

func (e *NebulaExporter) CollectMetrics(
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

		// TODO: uniform naming rules with _
		labels := []string{"nebula_cluster", "componentType", "instanceName"}
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

func (e *NebulaExporter) collect(wg *sync.WaitGroup, namespace, clusterName string, instance Instance, ch chan<- prometheus.Metric) {
	podIpAddress := instance.EndpointIP
	podHttpPort := instance.EndpointPort

	if podHttpPort <= 0 {
		return
	}

	klog.Infof("Collect %s:%s %s:%d Metrics ",
		clusterName, strings.ToUpper(instance.ComponentType),
		instance.Name, instance.EndpointPort)

	wg.Add(2)
	if instance.ComponentType == "storaged" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rocksDBStatus, err := getNebulaRocksDBStats(podIpAddress, podHttpPort)
			if err != nil {
				klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
				return
			}
			e.CollectMetrics(instance.Name, instance.ComponentType, namespace, clusterName, rocksDBStatus, ch)
		}()
	}

	go func() {
		defer wg.Done()
		metrics, err := getNebulaMetrics(podIpAddress, podHttpPort)
		if err != nil {
			klog.Errorf("get query metrics from %s:%d failed: %v", podIpAddress, podHttpPort, err)
			return
		}
		e.CollectMetrics(instance.Name, instance.ComponentType, namespace, clusterName, metrics, ch)
	}()

	go func() {
		defer wg.Done()
		statusMetrics := "count=1"
		if !isNebulaComponentRunning(podIpAddress, podHttpPort) {
			statusMetrics = "count=0"
		}
		e.CollectMetrics(instance.Name, instance.ComponentType, namespace, clusterName, []string{statusMetrics}, ch)
	}()
}

func (e *NebulaExporter) CollectFromStaticConfig(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, cluster := range e.Config.Clusters {
		cluster := cluster
		if cluster.Name == "" {
			cluster.Name = DefaultClusterName
		}
		for _, instance := range cluster.Instances {
			instance := instance
			if instance.Name == "" {
				instance.Name = fmt.Sprintf("%s-%s", instance.EndpointIP, instance.ComponentType)
			}
			e.collect(&wg, NonNamespace, cluster.Name, instance, ch)
		}
	}

	wg.Wait()
}

func (e *NebulaExporter) CollectFromKubernetes(ch chan<- prometheus.Metric) {
	listOpts := metav1.ListOptions{}

	listOpts.LabelSelector = e.Selector
	if listOpts.LabelSelector == "" && e.Cluster != "" {
		listOpts.LabelSelector = fmt.Sprintf("%s=%s", e.ClusterLabelKey, e.Cluster)
	}
	podLists, err := e.Client.CoreV1().Pods(e.Namespace).List(context.TODO(), listOpts)
	if err != nil {
		klog.Error(err)
		return
	}

	var wg sync.WaitGroup
	for _, item := range podLists.Items {
		pod := item

		clusterName, ok := pod.Labels[e.ClusterLabelKey]
		if !ok {
			continue
		}

		podIpAddress := pod.Status.PodIP
		for _, port := range pod.Spec.Containers[0].Ports {
			if port.ContainerPort == 0 {
				continue
			}
			var componentType string
			switch {
			case port.Name == e.GraphPortName:
				componentType = ComponentGraphdLabelVal
			case port.Name == e.MetaPortName:
				componentType = ComponentMetadLabelVal
			case port.Name == e.StoragePortName:
				componentType = ComponentStoragedLabelVal
			case port.Name == "http" && pod.Labels[ComponentLabelKey] != "":
				componentType = pod.Labels[ComponentLabelKey]
			}

			if componentType != "" {
				e.collect(&wg, pod.Namespace, clusterName, Instance{
					Name:          pod.Name,
					EndpointIP:    podIpAddress,
					EndpointPort:  port.ContainerPort,
					ComponentType: componentType,
				}, ch)
			}
		}
	}

	wg.Wait()
}

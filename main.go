package main

import (
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/vesoft-inc/nebula-stats-exporter/exporter"
)

var (
	restClient *kubernetes.Clientset
)

func main() {
	var (
		listenAddr = kingpin.Flag(
			"listen-address",
			"Address of nebula metrics server").
			Default(":9100").String()

		namespace = kingpin.Flag("namespace",
			"The namespace which nebula in").
			Default("default").String()

		cluster = kingpin.Flag("cluster",
			"The cluster name for nebula, default get metrics of all clusters in the namespace.").
			Default("").String()

		clusterLabelKey = kingpin.Flag("cluster-label-key",
			"The cluster name label key.").
			Default("").String()

		selector = kingpin.Flag("selector",
			"The selector (label query) to filter on pods.").
			Default("").String()

		graphPortName = kingpin.Flag("graph-port-name",
			"The graph port name of pod to collect metrics.").
			Default("http-graph").String()

		metaPortName = kingpin.Flag("meta-port-name",
			"The meta port name of pod to collect metrics.").
			Default("http-meta").String()

		storagePortName = kingpin.Flag("storage-port-name",
			"The storage port name of pod to collect metrics.").
			Default("http-storage").String()

		kubeconfig = kingpin.Flag("kube-config",
			"The kubernetes config file").
			Default("").String()

		maxRequest = kingpin.Flag("max-request",
			"Maximum number of parallel scrape requests. Use 0 to disable.").
			Default("40").Int()

		bareMetal = kingpin.Flag("bare-metal",
			"Whether running in bare metal environment").
			Default("false").Bool()

		bareMetalConfig = kingpin.Flag("bare-metal-config",
			"The bare metal config file").
			Default("/config.yaml").String()
	)

	kingpin.Version(version.Print("nebula-stats-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	nebulaExporter := &exporter.NebulaExporter{
		Namespace:       *namespace,
		Selector:        *selector,
		Cluster:         *cluster,
		ClusterLabelKey: *clusterLabelKey,
		GraphPortName:   *graphPortName,
		MetaPortName:    *metaPortName,
		StoragePortName: *storagePortName,
		ListenAddress:   *listenAddr,
	}

	if *bareMetal {
		raw, err := ioutil.ReadFile(*bareMetalConfig)
		if err != nil {
			klog.Fatalf("read config file failed: %v", err)
		}

		if err := yaml.Unmarshal(raw, &nebulaExporter.Config); err != nil {
			klog.Fatalf("unmarshal failed: %v", err)
		}

		if err := nebulaExporter.Config.Validate(); err != nil {
			klog.Fatalf("bare-metal config validation failed: %v", err)
		}
	} else {
		config, err := buildConfig(*kubeconfig)
		if err != nil {
			klog.Fatalf("build config failed: %v", err)
		}

		restClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			klog.Fatalf("create k8s client failed: %v", err)
		}

		nebulaExporter.Client = restClient
	}

	if err := nebulaExporter.Initialize(*maxRequest); err != nil {
		klog.Fatal(err)
	}

	klog.Infof("Providing metrics at %s/metrics", *listenAddr)
	klog.Fatal(http.ListenAndServe(*listenAddr, nebulaExporter))
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

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

	var nebulaExporter *exporter.NebulaExporter
	if *bareMetal {
		raw, err := ioutil.ReadFile(*bareMetalConfig)
		if err != nil {
			klog.Fatalf("read config file failed: %v", err)
		}

		config := exporter.StaticConfig{}
		if err := yaml.Unmarshal(raw, &config); err != nil {
			klog.Fatalf("unmarshal failed: %v", err)
		}

		nebulaExporter, err = exporter.NewNebulaExporter(*namespace, *cluster, *listenAddr, nil, config, *maxRequest)
		if err != nil {
			klog.Fatal(err)
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

		nebulaExporter, err = exporter.NewNebulaExporter(*namespace, *cluster, *listenAddr, restClient, exporter.StaticConfig{}, *maxRequest)
		if err != nil {
			klog.Fatal(err)
		}
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

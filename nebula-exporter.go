package main

import (
	"github.com/prometheus/common/version"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"

	. "github.com/nebula-stats-exporter/exporter"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var restClient *kubernetes.Clientset

func main() {
	var (
		listenAddr = kingpin.Flag(
			"listen-address",
			"Address of nebula metrics server").
			Default(":9100").String()

		namespace = kingpin.Flag("namespace",
			"the namespace whice nebula in").
			Default("default").String()

		maxRequest = kingpin.Flag("max-request",
			"Maximum number of parallel scrape requests. Use 0 to disable.").
			Default("40").Int()

		bareMetal = kingpin.Flag("bare-metal",
			"Barely Metal Static Config").
			Default("false").Bool()

		bareMetalConfig = kingpin.Flag("bare-metal-config-path",
			"Barely metal config file path.").
			Default("/config.yaml").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("nebula-stats-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var exporter *NebulaExporter
	if !*bareMetal {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Can't Create K8s Client: %v", err)
		}

		restClient, err = kubernetes.NewForConfig(config)

		if err != nil {
			log.Fatalf("Can't Create K8s Client: %v", err)
		}

		exporter, err = NewNebulaExporter(*namespace, *listenAddr, restClient, StaticConfig{}, *maxRequest)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		raw, err := ioutil.ReadFile(*bareMetalConfig)
		if err != nil {
			log.Fatal(err)
		}
		config := StaticConfig{}
		yaml.Unmarshal(raw, &config)

		exporter, err = NewNebulaExporter(*namespace, *listenAddr, nil, config, *maxRequest)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("Providing metrics at %s/metrics", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, exporter))
}

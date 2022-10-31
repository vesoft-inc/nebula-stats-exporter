<p align="center">
  <img src="https://nebula-website-cn.oss-cn-hangzhou.aliyuncs.com/nebula-website/images/nebulagraph-logo.png"/>
  <br> <a href="README.md">English</a> | <a href="README-CN.md">中文</a>
  <br>A distributed, scalable, lightning-fast graph database<br>
</p>

# Nebula stats exporter

[Nebula Graph](https://github.com/vesoft-inc/nebula-graph) exporter for Prometheus.

The metrics currently collected are:

- Graphd's metrics
- Graphd's space level metrics
- Metad's metrics
- Storaged's metrics
- Storaged's rocksdb metrics
- Metad Listener's metrics
- Storaged Listener's metrics
- Draienrd's metrics

## Building and running the exporter

```shell
$ git clone https://github.com/vesoft-inc/nebula-stats-exporter.git
$ cd nebula-stats-exporter
$ make build
$ ./nebula-stats-exporter --help
usage: nebula-stats-exporter [<flags>]

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --listen-address=":9100"  Address of nebula metrics server
      --namespace="default"     The namespace which nebula in
      --cluster=""              The cluster name for nebula, default get metrics of all clusters in the namespace.
      --kube-config=""          The kubernetes config file
      --max-request=40          Maximum number of parallel scrape requests. Use 0 to disable.
      --bare-metal              Whether running in bare metal environment
      --bare-metal-config="/config.yaml"
                                The bare metal config file
      --version                 Show application version.
```

The `--bare-metal-config` file examples:

```yaml
clusters:                                   # a list of clusters you want to monitor
  - name: nebula                            # the cluster name
    instances:                              # a list of instances for this cluster
      - name: metad0                        # the instance name
        endpointIP: 192.168.10.131          # the ip of this instance
        endpointPort: 19559                 # the port of this instance
        componentType: metad                # the component type of this instance, optional value metad, graphd and storaged.
      - ...
```

_See [config.yaml](deploy/bare-metal/config.yaml) for details._

## Basic Prometheus Configuration

Add a block to the `scrape_configs` of your prometheus.yaml config file:

```yaml
scrape_configs:
  - job_name: 'nebula-stats-exporter'
    static_configs:
      - targets: ['<<NEBULA_STATS_EXPORTER_HOSTNAME>>:9100'] # nebula-stats-exporter metrics endpoints
```

And adjust the host name `NEBULA_STATS_EXPORTER_HOSTNAME` accordingly.

## Run with Docker

```shell
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "<<PATH_OF_CONFIG_FILE>>:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.3.0 --bare-metal --bare-metal-config=/config.yaml

# For example:
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "$(pwd)/deploy/bare-metal/config.yaml:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.3.0 --bare-metal --bare-metal-config=/config.yaml
```

## Run on Bare Metal

[Here](deploy/bare-metal/README.md) is an introduction to deploy on Bare Metal.

## Run on Kubernetes

[Here](deploy/kubernetes/README.md) is an introduction to deploy on Kubernetes.

## Import grafana config

Please import [nebula-grafana.json](deploy/grafana/nebula-grafana.json) into your grafana.

Wait a while, then you can see the charts in grafana.

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

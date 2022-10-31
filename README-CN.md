<p align="center">
  <img src="https://nebula-website-cn.oss-cn-hangzhou.aliyuncs.com/nebula-website/images/nebulagraph-logo.png"/>
  <br> <a href="README.md">English</a> | <a href="README-CN.md">中文</a>
  <br>A distributed, scalable, lightning-fast graph database<br>
</p>

# Nebula stats exporter

[Nebula Graph](https://github.com/vesoft-inc/nebula-graph) exporter for Prometheus.

目前收集的 metrics 是：

- Graphd 的 metrics
- Graphd space 级别的 metrics
- Metad 的 metrics
- Storaged 的 metrics
- Stroaged 内置 rocksdb 的 metrics
- Metad Listener 的 metrics
- Storaged Listener 的 metrics
- Draienrd 的 metrics

## 构建运行 exporter

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

`--bare-metal-config` 配置文件例子:

```yaml
clusters:                                   # 您想要监控的 cluster 列表
  - name: nebula                            # cluster 的名称
    instances:                              # cluster 中 instance 列表
      - name: metad0                        # instance 的名称
        endpointIP: 192.168.10.131          # instance 的 ip
        endpointPort: 19559                 # instance 的端口
        componentType: metad                # instance 的组建类型, optional value metad, graphd and storaged.
      - ...
```

_详情请见 [config.yaml](deploy/bare-metal/config.yaml) 。_

## 基本的 Prometheus 配置

在 prometheus.yaml 配置文件的 `scrape_configs` 中添加一个配置块：

```yaml
scrape_configs:
  - job_name: 'nebula-stats-exporter'
    static_configs:
      - targets: ['<<NEBULA_STATS_EXPORTER_HOSTNAME>>:9100'] # nebula-stats-exporter metrics endpoints
```

并相应地调整主机名 `NEBULA_STATS_EXPORTER_HOSTNAME` 。

## 使用 Docker 运行

```shell
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "<<PATH_OF_CONFIG_FILE>>:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.3.0 --bare-metal --bare-metal-config=/config.yaml

# 例如:
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "$(pwd)/deploy/bare-metal/config.yaml:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.3.0 --bare-metal --bare-metal-config=/config.yaml
```

## 在裸机上运行

[这里](deploy/bare-metal/README-CN.md)有一个在裸机上部署的介绍。

## 在 Kubernetes 中运行

[这里](deploy/kubernetes/README-CN.md)有一个在 Kubernetes 中部署的介绍.

## 导入 grafana 配置

请将 [nebula-grafana.json](deploy/grafana/nebula-grafana.json) 导入到您的 grafana 中。

等待一会儿， 接下来您可以在 grafana 中看到图表了。

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

# 在裸机上运行样例

## 运行 nebula-stats-exporter

```shell
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "<<PATH_OF_CONFIG_FILE>>:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.1.0 --bare-metal --bare-metal-config=/config.yaml

# 例如:
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "$(pwd)/deploy/bare-metal/config.yaml:/config.yaml" \
    vesoft/nebula-stats-exporter:v3.1.0 --bare-metal --bare-metal-config=/config.yaml
```

## 安装 prometheus

您需要在 `prometheus.yaml` 文件中配置 nebula-stats-exporter 。 这里我们使用静态配置。 请在 `static_configs` 中为 nebula-stats-exporter 指定 metrics endpoints 。

下面是一个例子：

```yaml
global:
  scrape_interval:     5s
  evaluation_interval: 5s
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'nebula-stats-exporter'
    static_configs:
      - targets: ['<<NEBULA_STATS_EXPORTER_HOSTNAME>>:9100'] # nebula-stats-exporter metrics endpoints
```

并相应地调整主机名 `NEBULA_STATS_EXPORTER_HOSTNAME` 。然后运行 prometheus ：

```shell
$ docker run -d --restart=always --name=prometheus -p 9090:9090 \
    -v "<<PATH_OF_CONFIG_DIR>>:/etc/prometheus" \
    prom/prometheus

# For example:
$ docker run -d --restart=always --name=prometheus -p 9090:9090 \
    -v "$(pwd)/deploy/bare-metal/prometheus:/etc/prometheus" \
    prom/prometheus
```

## 安装 grafana

首先运行 grafana:

```shell
$ docker run -d --restart=always --name grafana -p 3000:3000 grafana/grafana
```

Grafana 的默认 `用户名/密码` 是 `admin/admin`, http 地址是 `http://127.0.0.1:3000`.

请添加前面安装的 `Prometheus` 数据源，并将 [nebula-grafana.json](../grafana/nebula-grafana.json) 导入到您的 grafana 中。

等待一会儿， 接下来您可以在 grafana 中看到图表了。

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

## 清理样例

```shell
$ docker rm -f grafana prometheus nebula-stats-exporter
```

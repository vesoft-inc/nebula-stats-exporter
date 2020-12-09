# 非 k8s 环境下部署监控使用文档

## 编写 config.yaml 文件

在 `config.yaml` 中的 `nebulaItems` 下添加监控的 nebula 组件，`instanceName` 用于标示组件名称，`endpointIP` 用于指定该组件的 IP 地址，`endpointPort` 用于指定该组件的 http 端口，`nebulaType` 用于指定该组件是那种类型的 nebula 组件。
例子如下：

```yaml
nebulaItems:
  - instanceName: metad-0
    endpointIP: 192.168.8.102 # nebula component ip address
    endpointPort: 12000
    nebulaType: metad
```

### 运行 nebula-stats-exporter

直接运行：

```bash
docker run -d --restart=always -p 9100:9100 -v {directory to config.yaml}:/config \
 vesoft/nebula-stats-exporter:v0.0.2  --bare-metal --bare-metal-config-path=/config/config.yaml
```

### 配置 prometheus

在 `prometheus.yaml` 中配置好 nebula-stats-exporter，这里需要使用静态配置，并在 `static_configs` 中指定 nebula-stats-exporter 的 metrics endpoints。

如下面的例子：

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
      - targets: ['192.168.0.103:9100'] # nebula-stats-exporter metrics endpoints
```

然后运行 prometheus：

```bash
docker run -d --name=prometheus --restart=always \
-p 9090:9090 -v {directory to prometheus config}:/etc/prometheus/ prom/prometheus
```

### 配置 grafana

先启动 grafana：

```bash
docker run -d -p 3000:3000 grafana/grafana
```

然后将 `deploy/grafana/bare-metal/nebula-grafana.json` 导入到 nebula 的 dashboard 中。

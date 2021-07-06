# Nebula Stats Exporter

<p align="center">
  <br>English | <a href="README-CN.md">中文</a>
</p>

## For K8s Deployment

Please refer to [charts/nebula-exporter/README.md](https://github.com/vesoft-inc/nebula-stats-exporter/tree/master/charts/nebula-exporter) 

## For Bare Metal Deployment(Non-K8s)

### How to configure nebula-stats-exporter

Add the nebula components under `nebulaItems` in `config.yaml`.
* `instanceName` is to identify the component name and `endpointIP` is to specify the IP address of the component.
* `endpointPort` is to specify the `http` port of the component.
* `componentType` is to specify the service type of the component: `metad`, `storaged`, or `graphd`.
* `endpointIP` is the host IP address if you are deploying nebula as `rpm`, or the container IP address if you are deploying nebula as docker.

The sample configuration file can be found in `deploy/bare-metal/config.yaml`

```yaml
nebulaItems:
  - instanceName: metad-0
    endpointIP: 192.168.8.102 # nebula component ip address
    endpointPort: 19559
    componentType: metad
```

### Startup of nebula-stats-exporter

Start arguments:

```bash
usage: nebula-stats-exporter [<flags>]

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --listen-address=":9100"  Address of nebula metrics server
      --namespace="default"     The namespace which nebula in
      --config="/root/.kube/config"  
                                The kubernetes config file
      --max-request=40          Maximum number of parallel scrape requests. Use 0 to disable.
      --bare-metal              Whether running in bare metal environment
      --bare-metal-config="/config.yaml"  
                                The bare metal config file
      --version                 Show application version.
```

Example:

The `-v` argument specifies the local directory to be mounted when docker is run: i.e. `-v /root/nebula:/config` will mount `/root/nebula` (where `config.yaml` is stored) to path `/config` of the contianer.

```bash
docker run \
    -d --restart=always \
    -p 9100:9100 \
    -v {absolute directory of config.yaml}:/config vesoft/nebula-stats-exporter:v0.0.3 \
    --bare-metal --bare-metal-config=/config/config.yaml
```

### Prometheus Configuration

It's required to configure nebula-stats-exporter's `metrics endpoint` in `prometheus.yml`, and here `static_configs` is used. There is a sample in `deploy/bare-metal/prometheus.yml` as well.

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

Run prometheus:

```bash
docker run \
    -d --name=prometheus --restart=always \
    -p 9090:9090 \
    -v {absolute directory of prometheus.yml}:/etc/prometheus/ prom/prometheus
```

### Grafana Configuration

* Run grafana:

```bash
docker run -d -p 3000:3000 grafana/grafana
```
* Add `Prometheus DataSource` in `grafana Data Sources`

* Import the [grafana.json](!https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/deploy/grafana/bare-metal/nebula-grafana.json) to the nebula dashboard.

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)
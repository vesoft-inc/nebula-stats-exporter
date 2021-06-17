# Deployment Monitoring for Non-k8s Environments

## Edit config.yaml File

You need to add the nebula monitor components under `nebulaItems` in the `config.yaml` file. `instanceName` marks the component name. `endpointIP`, `endpointPort` and `nebulaType` specify the component IP, HTTP port and type respectively.

Here is an example:

```yaml
nebulaItems:
  - instanceName: metad-0
    endpointIP: 127.0.0.1 # nebula host IP
    endpointPort: 19559
    componentType: metad
```

### Run nebula-stats-exporter

Startup parameters

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

Run directly:

```bash
docker run -d --restart=always -p 9100:9100 -v {directory to config.yaml}:/config \
 vesoft/nebula-stats-exporter:v0.0.3  --bare-metal --bare-metal-config=/config/config.yaml
```

### Configure Prometheus

You need to configure nebula-stats-exporter in the `prometheus.yml` file. Here we use the static configs. Please specify the metrics endpoints for nebula-stats-exporter in `static_configs`.

Here is an example:

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

Then run prometheus:

```bash
docker run -d --name=prometheus --restart=always \
-p 9090:9090 -v {directory to prometheus config}:/etc/prometheus/ prom/prometheus
```

### Configure grafana

First run grafana:

```bash
docker run -d -p 3000:3000 grafana/grafana
```

Then import the [nebula-grafana.json](../grafana/nebula-grafana.json) to nebula's dashboard.
s
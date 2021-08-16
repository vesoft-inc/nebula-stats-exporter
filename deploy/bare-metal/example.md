# Run on Bare Metal Example

## Run nebula-stats-exporter

```shell
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "<<PATH_OF_CONFIG_FILE>>:/config.yaml" \
    vesoft/nebula-stats-exporter:v0.0.5 --bare-metal --bare-metal-config=/config.yaml

# For example:
$ docker run -d --restart=always --name nebula-stats-exporter -p 9100:9100 \
    -v "$(pwd)/deploy/bare-metal/config.yaml:/config.yaml" \
    vesoft/nebula-stats-exporter:v0.0.5 --bare-metal --bare-metal-config=/config.yaml
```

## Install prometheus

You need to configure nebula-stats-exporter in the `prometheus.yaml` file. Here we use the static configs. Please specify the metrics endpoints for nebula-stats-exporter in `static_configs`.

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
      - targets: ['<<NEBULA_STATS_EXPORTER_HOSTNAME>>:9100'] # nebula-stats-exporter metrics endpoints
```

And adjust the host name `NEBULA_STATS_EXPORTER_HOSTNAME` accordingly. Then run prometheus:

```shell
$ docker run -d --restart=always --name=prometheus -p 9090:9090 \
    -v "<<PATH_OF_CONFIG_DIR>>:/etc/prometheus" \
    prom/prometheus

# For example:
$ docker run -d --restart=always --name=prometheus -p 9090:9090 \
    -v "$(pwd)/deploy/bare-metal/prometheus:/etc/prometheus" \
    prom/prometheus
```

## Install grafana

First run grafana:

```shell
$ docker run -d --restart=always --name grafana -p 3000:3000 grafana/grafana
```

The default `user/password` of grafana is `admin/admin`, and the http host is `http://127.0.0.1:3000`.

Please add `Prometheus` Data Sources, previously installed, in grafana,
and import [nebula-grafana.json](../grafana/nebula-grafana.json) into grafana.

Wait a while, then you can see the charts in grafana.

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

## Clean the example

```shell
$ docker rm -f grafana prometheus nebula-stats-exporter
```

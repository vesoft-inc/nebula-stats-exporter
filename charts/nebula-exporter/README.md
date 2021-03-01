# Prometheus Nebula Graph Exporter

A Prometheus exporter for Nebula Graph

Some of the metrics collections are:
- Graphd's query metrics
- Metad's heartbeat metrics
- Storaged's CRUD edge and vertex metrics

## Get Repo Info

```console
helm repo add nebula-exporter https://vesoft-inc.github.io/nebula-stats-exporter/charts
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

```console
# Helm 3
# helm install [NAME] [CHART] [flags]
$ helm install exporter nebula-exporter/nebula-exporter --version
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
# Helm 3
$ helm uninstall exporter
```

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/charts/nebula-exporter/values.yaml), or run these configuration commands:

```console
# Helm 3
$ helm show values nebula-exporter/nebula-exporter
```

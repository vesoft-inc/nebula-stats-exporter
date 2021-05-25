# Install Guide

### Requirements

* Kubernetes >= 1.12
* [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
* [Helm](https://helm.sh) >= 3.2.0

## Get Repo Info

```shell script
helm repo add nebula-exporter https://vesoft-inc.github.io/nebula-stats-exporter/charts
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Exporter

```shell script
# helm install [NAME] [CHART] [flags]
$ helm install nebula-exporter nebula-exporter/nebula-exporter --namespace=${namespace} --version=${chart_version}
```

Note:   
If the corresponding ${namespace} does not exist, you can create the namespace first by running the _kubectl create namespace ${namespace}_ command.

${chart_version} represents the chart version of Nebula Operator. For example, v0.2.0. You can view the currently supported versions by running the _helm search repo -l nebula-exporter_ command.

If you deploy nebula cluster by other tools rather than nebula-helm or nebula-operator, you need set the nebula component workload controller labels.

Eg: _kubectl -n ${namespace} label sts nebula-graphd  app.kubernetes.io/component=graphd_, the component type can be set to _graphd_、_metad_、_storaged_


_See [configuration](#configuration) below for custom demands._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/charts/nebula-exporter/values.yaml), or run these configuration commands:

```shell script
$ helm show values nebula-exporter/nebula-exporter
```

## Uninstall Exporter

```shell script
$ helm uninstall nebula-exporter
```


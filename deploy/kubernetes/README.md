# Run on Kubernetes Guide

### Requirements

* Kubernetes >= 1.16
* [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
* [Helm](https://helm.sh) >= 3.2.0

## Update Repo

```shell script
# If you have already added it, please skip.
$ helm repo add nebula-exporter https://vesoft-inc.github.io/nebula-stats-exporter/charts
$ helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install with helm

```shell
export NEBULA_EXPORTER_NAMESPACE=nebula     # the namespace you want to install the nebula exporter
export CHART_VERSION=v3.3.0                 # the chart version of Nebula Exporter

$ kubectl create namespace "${NEBULA_EXPORTER_NAMESPACE}" # If you have already created it, please skip.
$ helm install nebula-exporter nebula-exporter/nebula-exporter \
    --namespace="${NEBULA_EXPORTER_NAMESPACE}" \
    --version=${CHART_VERSION}

# Please wait a while for the cluster to be ready.
$ kubectl -n "${NEBULA_EXPORTER_NAMESPACE}" get pod -l "app.kubernetes.io/instance=nebula-exporter"
NAME                               READY   STATUS    RESTARTS   AGE
nebula-exporter-5964b765c9-4xfkm   1/1     Running   0          44s
```

Notes:

* `${chart_version}` represents the chart version of Nebula Exporter. For example, v3.3.0. You can view the currently supported versions by running the `helm search repo -l nebula-exporter` command.

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

### Upgrade with helm

```shell
$ helm upgrade nebula-exporter nebula-exporter/nebula-exporter \
    --namespace="${NEBULA_EXPORTER_NAMESPACE}" \
    --version=${CHART_VERSION} \
    --set replicaCount=2

# Please wait a while for the cluster to be ready.
$ kubectl -n "${NEBULA_EXPORTER_NAMESPACE}" get pod -l "app.kubernetes.io/instance=nebula-exporter"
NAME                               READY   STATUS    RESTARTS   AGE
nebula-exporter-5964b765c9-4xfkm   1/1     Running   0          44s
nebula-exporter-5964b765c9-t26n2   1/1     Running   0          83s
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

### Uninstall with helm

```shell
$ helm uninstall nebula-exporter --namespace="${NEBULA_EXPORTER_NAMESPACE}"
```

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Optional: chart parameters

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing).
To see all configurable options with detailed comments, visit the chart's [values.yaml](https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/charts/nebula-exporter/values.yaml),
or run the following command:

```shell
$ helm show values nebula-exporter/nebula-exporter
```

The following table lists is the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|:---------|:-----------|:-------|
| `nameOverride` | Override the name of the chart | `""` |
| `fullnameOverride` | Override the full name of the chart | `""` |
| `cluster` | Which Nebula Cluster's metrics to collect, default is all clusters in the same namespace | `""` |
| `replicaCount` | Nebula stats exporter replica number | `1` |
| `startUp.listenPort` | The nebula metrics server listening port | `9100` |
| `startUp.maxRequests` | Maximum number of parallel scrape requests, use 0 for no limit | `40` |
| `image.repository` | Nebula stats exporter image repository | `vesoft/nebula-stats-exporter` |
| `image.tag` | Nebula stats exporter image tag | `v3.3.0` |
| `image.pullPolicy` | Nebula stats exporter imagePullPolicy | `IfNotPresent` |
| `serviceAccount.create` | Specifies whether a service account should be created | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.name` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template | `nebula-exporter` |
| `serviceMonitor.enabled` | Whether to enable serviceMonitor | `false` |
| `serviceMonitor.prometheusServiceMatchLabels` | The prometheus ServiceMatchLabels, use `kubectl -n `${namespace}` get prometheus `${name}` -o=jsonpath='{.spec.serviceMonitorSelector.matchLabels}'` to find.| `{}` |
| `ingress.enabled` | Whether to enable ingress | `false` |
| `ingress.annotations` | The annotations for ingress | `{}` |
| `ingress.hosts` | The hosts for ingress | `chart-example.local` |
| `ingress.paths` | The paths for ingress | `["/metrics"]` |
| `ingress.tls` | The tls for ingress | `[]` |
| `ingress.ingressClass` | The ingressClass for ingress | `""` |
| `podAnnotations` | Nebula stats exporter pod annotations | `{}` |
| `podLabels` | Nebula stats exporter pod labels | `{}` |
| `livenessProbe` | Nebula stats exporter livenessProbe | `{"failureThreshold":2,"httpGet":{"path":"/health","scheme":"HTTP"},"initialDelaySeconds":30,"timeoutSeconds":10}` |
| `nodeSelector` | Nebula stats exporter nodeSelector | `{}` |
| `tolerations` | Nebula stats exporter tolerations | `{}` |
| `affinity` | Nebula stats exporter affinity | `{}` |

## Example

_See [example](example.md) for details._

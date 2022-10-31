# Run on Kubernetes Example

## Prerequisites

* A ready Kubernetes
* With [nebula-operator](https://github.com/vesoft-inc/nebula-operator) installed

## Install prometheus

```shell
$ kubectl create ns monitoring
$ helm -n monitoring install prometheus kube-prometheus-stack --repo https://prometheus-community.github.io/helm-charts
```

_See [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack) for details._

## Create several Nebula Clusters

```shell
$ export STORAGE_CLASS_NAME=gp2             # the storage class for the nebula cluster

$ kubectl create namespace example1
$ helm install nebula1 nebula-cluster --repo https://vesoft-inc.github.io/nebula-operator/charts \
    --namespace=example1 \
    --set nameOverride=nebula1 \
    --set nebula.storageClassName="${STORAGE_CLASS_NAME}" \
    --set nebula.graphd.replicas=1 \
    --set nebula.metad.replicas=1 \
    --set nebula.storaged.replicas=3
$ helm install nebula2 nebula-cluster --repo https://vesoft-inc.github.io/nebula-operator/charts \
    --namespace=example1 \
    --set nameOverride=nebula2 \
    --set nebula.storageClassName="${STORAGE_CLASS_NAME}" \
    --set nebula.graphd.replicas=1 \
    --set nebula.metad.replicas=3 \
    --set nebula.storaged.replicas=3

$ kubectl create namespace example2
$ helm install nebula3 nebula-cluster --repo https://vesoft-inc.github.io/nebula-operator/charts \
    --namespace=example2 \
    --set nameOverride=nebula3 \
    --set nebula.storageClassName="${STORAGE_CLASS_NAME}" \
    --set nebula.graphd.replicas=2 \
    --set nebula.metad.replicas=3 \
    --set nebula.storaged.replicas=3

$ helm install nebula4 nebula-cluster --repo https://vesoft-inc.github.io/nebula-operator/charts \
    --namespace=example2 \
    --set nameOverride=nebula4 \
    --set nebula.storageClassName="${STORAGE_CLASS_NAME}" \
    --set nebula.graphd.replicas=2 \
    --set nebula.metad.replicas=3 \
    --set nebula.storaged.replicas=5
```

_See [Install Nebula Cluster with helm](https://github.com/vesoft-inc/nebula-operator/blob/master/doc/user/nebula_cluster_helm_guide.md) for details._

## Install Nebula Exporter

```shell
$ helm install nebula-exporter nebula-exporter --repo https://vesoft-inc.github.io/nebula-stats-exporter/charts \
    --namespace=example1 --version=v3.3.0 \
    --set serviceMonitor.enabled=true \
    --set serviceMonitor.prometheusServiceMatchLabels.release=prometheus

$ helm install nebula-exporter nebula-exporter --repo https://vesoft-inc.github.io/nebula-stats-exporter/charts \
    --namespace=example2 --version=v3.3.0 \
    --set serviceMonitor.enabled=true \
    --set serviceMonitor.prometheusServiceMatchLabels.release=prometheus
```

## Import and view grafana

Forward the grafana service:

```shell
$ kubectl -n monitoring port-forward svc/prometheus-grafana 3000:80
```

The default `user/password` of grafana is `admin/prom-operator`, and the http host is `http://127.0.0.1:3000`.

Please import [nebula-grafana.json](../grafana/nebula-grafana.json) into grafana.

Wait a while, then you can see the charts in grafana.

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

## Clean the example

```shell
$ helm uninstall nebula-exporter --namespace=example1
$ helm uninstall nebula-exporter --namespace=example2
$ helm uninstall nebula1 --namespace=example1
$ helm uninstall nebula2 --namespace=example1
$ helm uninstall nebula3 --namespace=example2
$ helm uninstall nebula4 --namespace=example2
$ helm -n monitoring uninstall prometheus
$ kubectl delete crd alertmanagerconfigs.monitoring.coreos.com
$ kubectl delete crd alertmanagers.monitoring.coreos.com
$ kubectl delete crd podmonitors.monitoring.coreos.com
$ kubectl delete crd probes.monitoring.coreos.com
$ kubectl delete crd prometheuses.monitoring.coreos.com
$ kubectl delete crd prometheusrules.monitoring.coreos.com
$ kubectl delete crd servicemonitors.monitoring.coreos.com
$ kubectl delete crd thanosrulers.monitoring.coreos.com
$ kubectl delete ns monitoring example1 example1
```

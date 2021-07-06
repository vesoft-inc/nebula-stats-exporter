# 在 Kubernetes 中运行样例

## 依赖

* 就绪的 Kubernetes
* 安装了 [nebula-operator](https://github.com/vesoft-inc/nebula-operator)

## 安装 prometheus

```shell
$ kubectl create ns monitoring
$ helm -n monitoring install prometheus kube-prometheus-stack --repo https://prometheus-community.github.io/helm-charts
```

_详情请见 [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack) 。_

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

_详情请见 [Install Nebula Cluster with helm](https://github.com/vesoft-inc/nebula-operator/blob/master/doc/user/nebula_cluster_helm_guide.md) 。_

## Install Nebula Exporter

```shell
$ helm install nebula-exporter nebula-exporter --repo https://vesoft-inc.github.io/nebula-stats-exporter/charts \
    --namespace=example1 --version=v0.2.0 \
    --set serviceMonitor.enabled=true \
    --set serviceMonitor.prometheusServiceMatchLabels.release=prometheus

$ helm install nebula-exporter nebula-exporter --repo https://vesoft-inc.github.io/nebula-stats-exporter/charts \
    --namespace=example2 --version=v0.2.0 \
    --set serviceMonitor.enabled=true \
    --set serviceMonitor.prometheusServiceMatchLabels.release=prometheus
```

## 导入查看 grafana

转发 grafana 服务：

```shell
$ kubectl -n monitoring port-forward svc/prometheus-grafana 3000:80
```

Grafana 的默认 `用户名/密码` 是 `admin/prom-operator`, http 地址是 `http://127.0.0.1:3000`.

请将 [nebula-grafana.json](../grafana/nebula-grafana.json) 导入到您的 grafana 中。

等待一会儿， 接下来您可以在 grafana 中看到图表了。

![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

## 清理样例

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

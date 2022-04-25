# 在 Kubernetes 中运行指导

### 依赖

* Kubernetes >= 1.16
* [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
* [Helm](https://helm.sh) >= 3.2.0

## 更新 Repo

```shell script
# 如果您已经添加了，请跳过
$ helm repo add nebula-exporter https://vesoft-inc.github.io/nebula-stats-exporter/charts
$ helm repo update
```

_访问 [helm repo](https://helm.sh/docs/helm/helm_repo/) 来查看命令文档。_

## 使用 helm 安装

```shell
export NEBULA_EXPORTER_NAMESPACE=nebula     # 您想要安装 nebula exporter 的 namespace
export CHART_VERSION=v3.1.0                 # Nebula Exporter 的 chart 版本

$ kubectl create namespace "${NEBULA_EXPORTER_NAMESPACE}" # 如果您已经创建了，请跳过
$ helm install nebula-exporter nebula-exporter/nebula-exporter \
    --namespace="${NEBULA_EXPORTER_NAMESPACE}" \
    --version=${CHART_VERSION}

# 请等待 cluster 就绪。
$ kubectl -n "${NEBULA_EXPORTER_NAMESPACE}" get pod -l "app.kubernetes.io/instance=nebula-exporter"
NAME                               READY   STATUS    RESTARTS   AGE
nebula-exporter-5964b765c9-4xfkm   1/1     Running   0          44s
```

注意:

* `${chart_version}` 表示 Nebula Exporter chart 的版本。 例如，v3.1.0 。 您可以通过运行 `helm search repo -l nebula-exporter` 命令来查看当前支持的版本。

_访问 [helm install](https://helm.sh/docs/helm/helm_install/) 来查看命令文档。_

### 使用 helm 升级

```shell
$ helm upgrade nebula-exporter nebula-exporter/nebula-exporter \
    --namespace="${NEBULA_EXPORTER_NAMESPACE}" \
    --version=${CHART_VERSION} \
    --set replicaCount=2

# 请等待 cluster 就绪。
$ kubectl -n "${NEBULA_EXPORTER_NAMESPACE}" get pod -l "app.kubernetes.io/instance=nebula-exporter"
NAME                               READY   STATUS    RESTARTS   AGE
nebula-exporter-5964b765c9-4xfkm   1/1     Running   0          44s
nebula-exporter-5964b765c9-t26n2   1/1     Running   0          83s
```

_访问 [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) 来查看命令文档。_

### 使用 helm 卸载

```shell
$ helm uninstall nebula-exporter --namespace="${NEBULA_EXPORTER_NAMESPACE}"
```

_访问 [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) 来查看命令文档。_

## 可选: chart 参数

参考 [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing).
要查看带有详细注释的所有可配置选项，请访问  [values.yaml](https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/charts/nebula-exporter/values.yaml),
或者运行如下命令：

```shell
$ helm show values nebula-exporter/nebula-exporter
```

下表列出了图表的可配置参数及其默认值。

| 参数 | 描述 | 默认值 |
|:---------|:-----------|:-------|
| `nameOverride` | 重写 chart 的 name | `""` |
| `fullnameOverride` | 重写 chart 的 full name | `""` |
| `cluster` | 收集哪个 Nebula Cluster 的 metrics ，默认是同一个 namespace 下的所有集群 | `""` |
| `replicaCount` | Nebula stats exporter 的 replica 数 | `1` |
| `startUp.listenPort` | Nebula metrics 服务监听端口 | `9100` |
| `startUp.maxRequests` | 最大并行抓取请求数，使用 0 则不限制 | `40` |
| `image.repository` | Nebula stats exporter image repository | `vesoft/nebula-stats-exporter` |
| `image.tag` | Nebula stats exporter image tag | `v3.1.0` |
| `image.pullPolicy` | Nebula stats exporter imagePullPolicy | `IfNotPresent` |
| `serviceAccount.create` | 指定是否应创建 service account | `true` |
| `serviceAccount.annotations` | 添加到 service account 的 annotations | `{}` |
| `serviceAccount.name` | 要使用的 service account 的名称。 如果未设置且 create 为 true，则使用 fullname 模板生成名称 | `nebula-exporter` |
| `serviceMonitor.enabled` | 是否开启 serviceMonitor | `false` |
| `serviceMonitor.prometheusServiceMatchLabels` | Prometheus 的 ServiceMatchLabels, 使用 `kubectl -n `${namespace}` get prometheus `${name}` -o=jsonpath='{.spec.serviceMonitorSelector.matchLabels}'` 获取。| `{}` |
| `ingress.enabled` | 是否开启 ingress | `false` |
| `ingress.annotations` | Ingress 的 annotations | `{}` |
| `ingress.hosts` | Ingress 的 hosts | `chart-example.local` |
| `ingress.paths` | Ingress 的 paths | `["/metrics"]` |
| `ingress.tls` | Ingress 的 tls | `[]` |
| `ingress.ingressClass` | Ingress 的 ingressClass | `""` |
| `podAnnotations` | Nebula stats exporter pod annotations | `{}` |
| `podLabels` | Nebula stats exporter pod labels | `{}` |
| `livenessProbe` | Nebula stats exporter livenessProbe | `{"failureThreshold":2,"httpGet":{"path":"/health","scheme":"HTTP"},"initialDelaySeconds":30,"timeoutSeconds":10}` |
| `nodeSelector` | Nebula stats exporter nodeSelector | `{}` |
| `tolerations` | Nebula stats exporter tolerations | `{}` |
| `affinity` | Nebula stats exporter affinity | `{}` |

## Example

_详情请见 [样例](example-CN.md) 。_

# 在裸机上运行

## 配置 `config.yaml` 文件

您需要在 `config.yaml` 文件的 `clusters` 下添加您要监控的 Nebula Cluster 。

现在支持监控多个 Nebula Cluster 。

例如：

```yaml
clusters:                                   # 您想要监控的 cluster 列表
  - name: nebula                            # cluster 的名称
    instances:                              # cluster 中 instance 列表
      - name: metad0                        # instance 的名称
        endpointIP: 192.168.10.131          # instance 的 ip
        endpointPort: 19559                 # instance 的端口
        componentType: metad                # instance 的组建类型, optional value metad, graphd and storaged.
      - ...
# 废弃: 使用 clusters 来替换
nebulaItems:                                # 类似 clusters/instances, 默认的 cluster 名称为 '_nebula'
  - instanceName: metad0                    # 类似 clusters/instances/name
    endpointIP: 192.168.10.131              # 类似 clusters/instances/endpointIP
    endpointPort: 19559                     # 类似 clusters/instances/endpointPort
    componentType: metad                    # 类似 clusters/instances/componentType
  - ...
```

_详情请见 [config.yaml](config.yaml) 。_

## 样例

_详情请见 [样例](example-CN.md) 。_

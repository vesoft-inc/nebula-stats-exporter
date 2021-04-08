# 非 k8s 环境下部署监控使用文档

### 配置nebula-stats-exporter

在config.yaml中的nebulaItems下添加监控的nebula组件，instanceName用于标识组件名称，endpointIP用于指定该组件的 IP 地址，
endpointPort用于指定该组件的 http 端口，nebulaType用于指定该组件的服务类型: metad、storaged 或者 graphd.  
如果你是rpm方式部署nebula，那么endpointIP请填写主机IP地址，如果你是docker方式部署nebula，那么endpointIP请填写容器IP地址.
  
配置示例文件可参考deploy/bare-metal/config.yaml

```yaml
nebulaItems:
  - instanceName: metad-0
    endpointIP: 192.168.8.102 # nebula component ip address
    endpointPort: 12000
    nebulaType: metad
```

### 运行 nebula-stats-exporter

直接运行： 

-v参数指定docker运行时需要挂载的本地目录，-v /root/nebula:/config，将存放config.yaml的/root/nebula挂载到容器/config下.

```bash
docker run -d --network=host --restart=always -p 9100:9100 -v {absolute directory of config.yaml}:/config \
 vesoft/nebula-stats-exporter:v0.0.2  --bare-metal --bare-metal-config-path=/config/config.yaml
```

### 配置 prometheus

在prometheus.yaml中配置好nebula-stats-exporter的metrics endpoint，这里需要使用静态配置.
 
配置示例文件可参考deploy/bare-metal/prometheus.yaml

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

然后运行 prometheus：

```bash
docker run -d --name=prometheus --restart=always \
-p 9090:9090 -v {absolute directory of prometheus.yaml}:/etc/prometheus/ prom/prometheus
```

### 配置 grafana

先启动 grafana：

```bash
docker run -d -p 3000:3000 grafana/grafana
```

然后将 [grafana.json](!https://github.com/vesoft-inc/nebula-stats-exporter/blob/master/deploy/grafana/bare-metal/nebula-grafana.json) 
导入到nebula的dashboard中。
    
![](https://user-images.githubusercontent.com/51590253/84129424-860abb80-aa74-11ea-9208-c5a66cade0f8.gif)

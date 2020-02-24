# 非k8s环境下部署监控使用文档

### 编写config.yaml文件
需要在config.yaml中的nebulaItems下添加监控的nebula组件，instanceName用于标示
组件名称，endpointIP用于指定该组件的ip地址，endpointPort用于指定该组件的http端口，
nebulaType用于指定该组件是那种类型的nebula组件。
例子如下：
```yaml
nebulaItems:
  - instanceName: metad-0
    endpointIP: 127.0.0.1
    endpointPort: 12000
    nebulaType: metad
```

### 运行nebula-stats-exporter 
直接运行：
```bash
docker run -d --restart=always -p 9100:9100 -v {path to config.yaml}:/config \
 vesoft/nebula-stats-exporter:v0.0.1  --bare-metal --bare-metal-config-path=/config/config.yaml 
```

### 配置prometheus 
需要在prometheus.yaml中配置好nebula-stats-exporter,这里需要使用静态配置,
需要在static_configs中指定nebula-stats-exporter的metrics endpoints.

如下面的例子：
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
然后运行prometheus:
```bash
docker run -d --name=prometheus --restart=always \
-p 9090:9090 -v {path to promethues config}:/etc/prometheus/ prom/prometheus
```

### 配置grafana 

先启动grafana：
```bash
docker run -d -p 3000:3000 grafana/grafana
```
然后将deploy/grafana/bare-metal/nebula-grafana.json导入到nebula的dashboard中。

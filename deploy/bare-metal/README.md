# Run on Bare Metal

## Configure `config.yaml` file

You need to add the Nebula Cluster, which you want to monitor, under `clusters` in the `config.yaml` file.

Now support to monitor multi Nebula Clusters.

For example:

```yaml
clusters:                                   # a list of clusters you want to monitor
  - name: nebula                            # the cluster name
    instances:                              # a list of instances for this cluster
      - name: metad0                        # the instance name
        endpointIP: 192.168.10.131          # the ip of this instance
        endpointPort: 19559                 # the port of this instance
        componentType: metad                # the component type of this instance, optional value metad, graphd and storaged.
      - ...
```

_See [config.yaml](config.yaml) for details._

## Example

_See [example](example.md) for details._

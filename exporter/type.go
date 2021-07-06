package exporter

const NebulaItemClusterName = "_nebula"

type (
	StaticConfig struct {
		Version     string       `yaml:"version"`
		Clusters    []Cluster    `yaml:"clusters"`
		// Deprecated: use Clusters instead.
		// TODO: Remove NebulaItems
		NebulaItems []NebulaItem `yaml:"nebulaItems"`
	}

	Cluster struct {
		Name      string     `yaml:"name"`
		Instances []Instance `yaml:"instances"`
	}

	Instance struct {
		Name          string `yaml:"name"`
		EndpointIP    string `yaml:"endpointIP"`
		EndpointPort  int32  `yaml:"endpointPort"`
		ComponentType string `yaml:"componentType"`
	}

	NebulaItem struct {
		InstanceName  string `yaml:"instanceName"`
		EndpointIP    string `yaml:"endpointIP"`
		EndpointPort  int32  `yaml:"endpointPort"`
		ComponentType string `yaml:"componentType"`
	}
)

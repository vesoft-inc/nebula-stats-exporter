package exporter

type StaticConfig struct {
	Version     string       `yaml:"version"`
	NebulaItems []NebulaItem `yaml:"nebulaItems"`
}

type NebulaItem struct {
	InstanceName  string `yaml:"instanceName"`
	EndpointIP    string `yaml:"endpointIP"`
	EndpointPort  int32  `yaml:"endpointPort"`
	ComponentType string `yaml:"componentType"`
}

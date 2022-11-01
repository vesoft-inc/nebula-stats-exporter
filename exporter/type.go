package exporter

import "fmt"

const (
	DefaultClusterName = "default"

	ComponentLabelKey = "app.kubernetes.io/component"
	ClusterLabelKey   = "app.kubernetes.io/cluster"

	ComponentGraphdLabelVal   = "graphd"
	ComponentMetadLabelVal    = "metad"
	ComponentStoragedLabelVal = "storaged"

	// FQNamespace represents the prometheus FQName
	FQNamespace  = "nebula"
	NonNamespace = "none_namespace"

	ComponentTypeGraphd          = "graphd"
	ComponentTypeMetad           = "metad"
	ComponentTypeStoraged        = "storaged"
	ComponentTypeMetaListener    = "metad-listener"
	ComponentTypeStorageListener = "storaged-listener"
	ComponentTypeDrainer         = "drainerd"
)

type (
	StaticConfig struct {
		Version  string    `yaml:"version"`
		Clusters []Cluster `yaml:"clusters"`
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
)

func (s *StaticConfig) Validate() error {
	if len(s.Clusters) == 0 {
		return fmt.Errorf("at least one cluster in config")
	}
	for _, cluster := range s.Clusters {
		for _, instance := range cluster.Instances {
			switch instance.ComponentType {
			case ComponentTypeGraphd, ComponentTypeMetad, ComponentTypeStoraged, ComponentTypeMetaListener, ComponentTypeStorageListener, ComponentTypeDrainer:
				continue
			default:
				return fmt.Errorf("invalid component type: %s", instance.ComponentType)
			}
		}

	}

	return nil
}

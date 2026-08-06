package cluster

import (
	"encoding/json"
	"fmt"
	"os"
)

// Node describes one physical node's identity and network address.
type Node struct {
	ID       string `json:"id"`
	HTTPAddr string `json:"http_addr"` // e.g. "localhost:8001"
}

// Config is the static list of nodes making up the cluster.
type Config struct {
	Nodes []Node `json:"nodes"`
}

// LoadConfig reads a cluster config JSON file from disk.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cluster: read config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cluster: parse config %s: %w", path, err)
	}
	if len(cfg.Nodes) == 0 {
		return nil, fmt.Errorf("cluster: config %s defines no nodes", path)
	}
	return &cfg, nil
}

// AddrOf returns the HTTP address for a given node ID, or false if unknown.
func (c *Config) AddrOf(id string) (string, bool) {
	for _, n := range c.Nodes {
		if n.ID == id {
			return n.HTTPAddr, true
		}
	}
	return "", false
}

// IDs returns every node ID in the config, in file order.
func (c *Config) IDs() []string {
	ids := make([]string, len(c.Nodes))
	for i, n := range c.Nodes {
		ids[i] = n.ID
	}
	return ids
}

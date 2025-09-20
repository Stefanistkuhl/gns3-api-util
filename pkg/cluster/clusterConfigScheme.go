package cluster

type Config struct {
	Settings Settings  `toml:"settings"`
	Clusters []Cluster `toml:"cluster"`
}

type Settings struct {
	DefaultMaxGroups int    `toml:"default_max_groups"`
	DefaultProtocol  string `toml:"default_protocol"`
}

type Cluster struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Nodes       []Node `toml:"node"`
}

type Node struct {
	Host      string `toml:"host"`
	Port      int    `toml:"port"`
	User      string `toml:"user"`
	Protocol  string `toml:"protocol"`
	Weight    int    `toml:"weight"`
	MaxGroups int    `toml:"max_groups"`
}

func NewConfig() Config {
	var cfg Config
	cfg.Settings.DefaultMaxGroups = 3
	cfg.Settings.DefaultProtocol = "http"
	return cfg
}

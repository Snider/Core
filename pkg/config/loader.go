package config

import (
	"path/filepath"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
	"gopkg.in/yaml.v3"
)

// Load reads a YAML configuration file from the given medium and path.
// Returns the parsed data as a map, or an error if the file cannot be read or parsed.
func Load(m io.Medium, path string) (map[string]any, error) {
	content, err := m.Read(path)
	if err != nil {
		return nil, core.E("config.Load", "failed to read config file: "+path, err)
	}

	data := make(map[string]any)
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, core.E("config.Load", "failed to parse config file: "+path, err)
	}

	return data, nil
}

// Save writes configuration data to a YAML file at the given path.
// It ensures the parent directory exists before writing.
func Save(m io.Medium, path string, data map[string]any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return core.E("config.Save", "failed to marshal config", err)
	}

	dir := filepath.Dir(path)
	if err := m.EnsureDir(dir); err != nil {
		return core.E("config.Save", "failed to create config directory: "+dir, err)
	}

	if err := m.Write(path, string(out)); err != nil {
		return core.E("config.Save", "failed to write config file: "+path, err)
	}

	return nil
}

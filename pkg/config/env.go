package config

import (
	"os"
	"strings"
)

// LoadEnv parses environment variables with the given prefix and returns
// them as a flat map with dot-notation keys.
//
// For example, with prefix "CORE_CONFIG_":
//
//	CORE_CONFIG_FOO_BAR=baz  ->  {"foo.bar": "baz"}
//	CORE_CONFIG_EDITOR=vim   ->  {"editor": "vim"}
func LoadEnv(prefix string) map[string]any {
	result := make(map[string]any)

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]

		// Strip prefix and convert to dot notation
		key := strings.TrimPrefix(name, prefix)
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, "_", ".")

		result[key] = value
	}

	return result
}

package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

// ConfigFormat defines an interface for loading and saving configuration data in
// various formats. Each format implementation is responsible for serializing and
// deserializing data between a file and a map of key-value pairs.
type ConfigFormat interface {
	// Load reads data from the specified path and returns it as a map.
	Load(path string) (map[string]interface{}, error)
	// Save writes the provided data map to the specified path.
	Save(path string, data map[string]interface{}) error
}

// JSONFormat implements the ConfigFormat interface for JSON files. It provides
// methods to read from and write to files in JSON format.
type JSONFormat struct{}

// Load reads a JSON file from the given path and decodes it into a map.
// The keys of the map are strings, and the values are of type interface{}.
func (f *JSONFormat) Load(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Save encodes the provided map into JSON format and writes it to the given
// path. The output is indented for readability.
func (f *JSONFormat) Save(path string, data map[string]interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// YAMLFormat implements the ConfigFormat interface for YAML files. It provides
// methods to read from and write to files in YAML format.
type YAMLFormat struct{}

// Load reads a YAML file from the given path and decodes it into a map.
func (f *YAMLFormat) Load(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Save encodes the provided map into YAML format and writes it to the given
// path.
func (f *YAMLFormat) Save(path string, data map[string]interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, yamlData, 0644)
}

// INIFormat implements the ConfigFormat interface for INI files. It handles
// the structured format of INI files, including sections and keys.
type INIFormat struct{}

// Load reads an INI file and converts its sections and keys into a single map.
// Keys in the map are formed by concatenating the section name and key name with
// a dot (e.g., "section.key").
func (f *INIFormat) Load(path string) (map[string]interface{}, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			result[section.Name()+"."+key.Name()] = key.Value()
		}
	}
	return result, nil
}

// Save writes a map of key-value pairs to an INI file. Keys in the map are
// split by a dot to determine the section and key for the INI file.
func (f *INIFormat) Save(path string, data map[string]interface{}) error {
	cfg := ini.Empty()
	for key, value := range data {
		parts := strings.SplitN(key, ".", 2)
		section := ini.DefaultSection
		keyName := parts[0]
		if len(parts) > 1 {
			section = parts[0]
			keyName = parts[1]
		}
		if _, err := cfg.Section(section).NewKey(keyName, fmt.Sprintf("%v", value)); err != nil {
			return err
		}
	}
	return cfg.SaveTo(path)
}

// XMLFormat implements the ConfigFormat interface for XML files. It uses a
// simple structure with a root "config" element containing a series of "entry"
// elements, each with a "key" and "value".
type XMLFormat struct{}

// xmlEntry is a helper struct for marshaling and unmarshaling XML data.
type xmlEntry struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

// Load reads an XML file and parses it into a map. It expects the XML to have
// a specific structure as defined by the xmlEntry struct.
func (f *XMLFormat) Load(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v struct {
		XMLName xml.Name   `xml:"config"`
		Entries []xmlEntry `xml:"entry"`
	}
	if err := xml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for _, entry := range v.Entries {
		result[entry.Key] = entry.Value
	}
	return result, nil
}

// Save writes a map of key-value pairs to an XML file. The data is structured
// with a root "config" element and child "entry" elements.
func (f *XMLFormat) Save(path string, data map[string]interface{}) error {
	var entries []xmlEntry
	for key, value := range data {
		entries = append(entries, xmlEntry{
			Key:   key,
			Value: fmt.Sprintf("%v", value),
		})
	}
	xmlData, err := xml.MarshalIndent(struct {
		XMLName xml.Name   `xml:"config"`
		Entries []xmlEntry `xml:"entry"`
	}{Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, xmlData, 0644)
}

// GetConfigFormat returns a ConfigFormat implementation based on the file
// extension of the provided path. This allows the config service to dynamically
// handle different file formats.
//
// Example:
//
//	format, err := GetConfigFormat("settings.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// format is now a JSONFormat
func GetConfigFormat(path string) (ConfigFormat, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return &JSONFormat{}, nil
	case ".yaml", ".yml":
		return &YAMLFormat{}, nil
	case ".ini":
		return &INIFormat{}, nil
	case ".xml":
		return &XMLFormat{}, nil
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}
}

// SaveKeyValues saves a map of key-value pairs to a file in the config
// directory. The file format is determined by the extension of the `key`
// parameter. This method is a convenient way to store structured data in a
// format of choice.
//
// Example:
//
//	data := map[string]interface{}{"host": "localhost", "port": 8080}
//	err := cfg.SaveKeyValues("database.yml", data)
//	if err != nil {
//		log.Printf("Error saving database config: %v", err)
//	}
func (s *Service) SaveKeyValues(key string, data map[string]interface{}) error {
	format, err := GetConfigFormat(key)
	if err != nil {
		return err
	}
	filePath := filepath.Join(s.ConfigDir, key)
	return format.Save(filePath, data)
}

// LoadKeyValues loads a map of key-value pairs from a file in the config
// directory. The file format is determined by the extension of the `key`
// parameter. This allows for easy retrieval of data stored in various formats.
//
// Example:
//
//	dbConfig, err := cfg.LoadKeyValues("database.yml")
//	if err != nil {
//		log.Printf("Error loading database config: %v", err)
//	}
//	port, ok := dbConfig["port"].(int)
//	// ...
func (s *Service) LoadKeyValues(key string) (map[string]interface{}, error) {
	format, err := GetConfigFormat(key)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(s.ConfigDir, key)
	return format.Load(filePath)
}

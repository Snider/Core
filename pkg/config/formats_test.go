package config

import (
	"os"
	"reflect"
	"testing"
)

func TestConfigFormats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := &Service{
		ConfigDir: tempDir,
	}

	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123.0,
		"key3": true,
	}

	testCases := []struct {
		format   string
		filename string
	}{
		{"json", "test.json"},
		{"yaml", "test.yaml"},
		{"ini", "test.ini"},
		{"xml", "test.xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			// Test SaveKeyValues
			err := service.SaveKeyValues(tc.filename, testData)
			if err != nil {
				t.Fatalf("SaveKeyValues failed for %s: %v", tc.format, err)
			}

			// Test LoadKeyValues
			loadedData, err := service.LoadKeyValues(tc.filename)
			if err != nil {
				t.Fatalf("LoadKeyValues failed for %s: %v", tc.format, err)
			}

			// INI format saves everything as strings, so we need to adjust the expected data
			expectedData := testData
			if tc.format == "ini" {
				expectedData = map[string]interface{}{
					"DEFAULT.key1": "value1",
					"DEFAULT.key2": "123",
					"DEFAULT.key3": "true",
				}
			}

			if tc.format == "yaml" {
				// The yaml library unmarshals numbers as int if they don't have a decimal point.
				if val, ok := loadedData["key2"].(int); ok {
					loadedData["key2"] = float64(val)
				}
			}

			if tc.format == "xml" {
				expectedData = map[string]interface{}{
					"key1": "value1",
					"key2": "123",
					"key3": "true",
				}
			}

			if !reflect.DeepEqual(expectedData, loadedData) {
				t.Errorf("Loaded data does not match original data for %s.\nExpected: %v\nGot: %v", tc.format, expectedData, loadedData)
			}
		})
	}
}

func TestGetConfigFormat(t *testing.T) {
	testCases := []struct {
		filename     string
		expectedType interface{}
		expectError  bool
	}{
		{"config.json", &JSONFormat{}, false},
		{"config.yaml", &YAMLFormat{}, false},
		{"config.yml", &YAMLFormat{}, false},
		{"config.ini", &INIFormat{}, false},
		{"config.xml", &XMLFormat{}, false},
		{"config.txt", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			format, err := GetConfigFormat(tc.filename)
			if (err != nil) != tc.expectError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if !tc.expectError && reflect.TypeOf(format) != reflect.TypeOf(tc.expectedType) {
				t.Errorf("Expected format type %T, got %T", tc.expectedType, format)
			}
		})
	}
}

func TestFormatLoadErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-format-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("JSON Load with non-existent file", func(t *testing.T) {
		format := &JSONFormat{}
		_, err := format.Load(tempDir + "/nonexistent.json")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("JSON Load with invalid JSON", func(t *testing.T) {
		format := &JSONFormat{}
		invalidPath := tempDir + "/invalid.json"
		if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}
		_, err := format.Load(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("YAML Load with non-existent file", func(t *testing.T) {
		format := &YAMLFormat{}
		_, err := format.Load(tempDir + "/nonexistent.yaml")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("INI Load with non-existent file", func(t *testing.T) {
		format := &INIFormat{}
		_, err := format.Load(tempDir + "/nonexistent.ini")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("XML Load with non-existent file", func(t *testing.T) {
		format := &XMLFormat{}
		_, err := format.Load(tempDir + "/nonexistent.xml")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("XML Load with invalid XML", func(t *testing.T) {
		format := &XMLFormat{}
		invalidPath := tempDir + "/invalid.xml"
		if err := os.WriteFile(invalidPath, []byte("not valid xml <><>"), 0644); err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}
		_, err := format.Load(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid XML")
		}
	})

	t.Run("YAML Load with invalid YAML", func(t *testing.T) {
		format := &YAMLFormat{}
		invalidPath := tempDir + "/invalid.yaml"
		// Tabs in YAML cause errors
		if err := os.WriteFile(invalidPath, []byte("key:\n\t- invalid indent"), 0644); err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}
		_, err := format.Load(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestSaveKeyValuesErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-save-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := &Service{
		ConfigDir: tempDir,
	}

	t.Run("SaveKeyValues with unsupported format", func(t *testing.T) {
		err := service.SaveKeyValues("test.txt", map[string]interface{}{"key": "value"})
		if err == nil {
			t.Error("Expected error for unsupported format")
		}
	})

	t.Run("LoadKeyValues with unsupported format", func(t *testing.T) {
		_, err := service.LoadKeyValues("test.txt")
		if err == nil {
			t.Error("Expected error for unsupported format")
		}
	})

	t.Run("LoadKeyValues with non-existent file", func(t *testing.T) {
		_, err := service.LoadKeyValues("nonexistent.json")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

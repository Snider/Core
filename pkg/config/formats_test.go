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
		filename      string
		expectedType  interface{}
		expectError   bool
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

package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiff_Good_NoBreaking(t *testing.T) {
	tmpDir := t.TempDir()

	baseSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	revSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.1.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
  /status:
    get:
      operationId: getStatus
      responses:
        "200":
          description: OK
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	revPath := filepath.Join(tmpDir, "rev.yaml")
	_ = os.WriteFile(basePath, []byte(baseSpec), 0644)
	_ = os.WriteFile(revPath, []byte(revSpec), 0644)

	result, err := Diff(basePath, revPath)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if result.Breaking {
		t.Error("expected no breaking changes for adding endpoint")
	}
}

func TestDiff_Good_Breaking(t *testing.T) {
	tmpDir := t.TempDir()

	baseSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
  /users:
    get:
      operationId: getUsers
      responses:
        "200":
          description: OK
`
	revSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "2.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	revPath := filepath.Join(tmpDir, "rev.yaml")
	_ = os.WriteFile(basePath, []byte(baseSpec), 0644)
	_ = os.WriteFile(revPath, []byte(revSpec), 0644)

	result, err := Diff(basePath, revPath)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !result.Breaking {
		t.Error("expected breaking change for removed endpoint")
	}
}

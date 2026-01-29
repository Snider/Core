# SDK Build Examples

## Generate All SDKs

```bash
core build sdk
```

## Specific Language

```bash
core build sdk --lang typescript
core build sdk --lang php
core build sdk --lang go
```

## Custom Spec

```bash
core build sdk --spec ./api/openapi.yaml
```

## With Version

```bash
core build sdk --version v2.0.0
```

## Preview

```bash
core build sdk --dry-run
```

## Configuration

`.core/sdk.yaml`:

```yaml
version: 1

spec: ./api/openapi.yaml

languages:
  - name: typescript
    output: sdk/typescript
    package: "@myorg/api-client"

  - name: php
    output: sdk/php
    namespace: MyOrg\ApiClient

  - name: go
    output: sdk/go
    module: github.com/myorg/api-client-go
```

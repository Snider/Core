# SDK Examples

## Validate

```bash
core sdk validate
core sdk validate --spec ./api.yaml
```

## Diff

```bash
# Compare with tag
core sdk diff --base v1.0.0

# Compare files
core sdk diff --base ./old-api.yaml --spec ./new-api.yaml
```

## Output

```
Breaking changes detected:

- DELETE /users/{id}/profile
  Endpoint removed

- PATCH /users/{id}
  Required field 'email' added

Non-breaking changes:

+ POST /users/{id}/avatar
  New endpoint added
```

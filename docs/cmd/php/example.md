# PHP Examples

## Development

```bash
# Start all services
core php dev

# With HTTPS
core php dev --https

# Skip services
core php dev --no-vite --no-horizon
```

## Testing

```bash
# Run all
core php test

# Parallel
core php test --parallel

# With coverage
core php test --coverage

# Filter
core php test --filter UserTest
```

## Code Quality

```bash
# Format
core php fmt --fix

# Static analysis
core php analyse --level 9
```

## Deployment

```bash
# Production
core php deploy

# Staging
core php deploy --staging

# Wait for completion
core php deploy --wait

# Check status
core php deploy:status

# Rollback
core php deploy:rollback
```

## Configuration

### .env

```env
COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=your-api-token
COOLIFY_APP_ID=production-app-id
COOLIFY_STAGING_APP_ID=staging-app-id
```

### .core/php.yaml

```yaml
version: 1

dev:
  domain: myapp.test
  ssl: true
  services:
    - frankenphp
    - vite
    - horizon
    - reverb
    - redis

deploy:
  coolify:
    server: https://coolify.example.com
    project: my-project
```

## Package Linking

```bash
# Link local packages
core php packages link ../my-package

# Update linked
core php packages update

# Unlink
core php packages unlink my-package
```

## SSL Setup

```bash
core php ssl
core php ssl --domain myapp.test
```

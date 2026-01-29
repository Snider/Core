# core php

Laravel/PHP development environment with FrankenPHP, Vite, Horizon, Reverb, and Redis.

## Commands

| Command | Description |
|---------|-------------|
| `core php dev` | Start development environment |
| `core php test` | Run PHPUnit/Pest tests |
| `core php fmt` | Format with Laravel Pint |
| `core php analyse` | Static analysis with PHPStan |
| `core php build` | Build production container |
| `core php deploy` | Deploy to Coolify |

## Development Environment

```bash
# Start all services
core php dev
```

This starts:
- FrankenPHP/Octane (HTTP server)
- Vite dev server (frontend)
- Laravel Horizon (queues)
- Laravel Reverb (WebSockets)
- Redis

```bash
# View unified logs
core php logs

# Stop all services
core php stop
```

## Testing

```bash
# Run tests
core php test

# Parallel testing
core php test --parallel

# With coverage
core php test --coverage
```

## Code Quality

```bash
# Format code
core php fmt

# Static analysis
core php analyse

# Run both
core php fmt && core php analyse
```

## Building

```bash
# Build Docker container
core php build

# Build LinuxKit image
core php build --type linuxkit

# Run production locally
core php serve --production
```

## Deployment

```bash
# Deploy to Coolify
core php deploy

# Deploy to staging
core php deploy --staging

# Check deployment status
core php deploy:status

# Rollback
core php deploy:rollback
```

## Package Management

Link local packages for development:

```bash
# Link a local package
core php packages link ../my-package

# Update linked packages
core php packages update

# Unlink
core php packages unlink my-package
```

## SSL/HTTPS

Local SSL with mkcert:

```bash
# Auto-configured with core php dev
# Uses mkcert for trusted local certificates
```

## Configuration

Optional `.core/php.yaml`:

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

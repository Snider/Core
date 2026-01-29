# core php

Laravel/PHP development tools with FrankenPHP.

## Commands

| Command | Description |
|---------|-------------|
| `dev` | Start development environment |
| `logs` | View service logs |
| `stop` | Stop all services |
| `status` | Show service status |
| `ssl` | Setup SSL certificates with mkcert |
| `build` | Build Docker or LinuxKit image |
| `serve` | Run production container |
| `shell` | Open shell in running container |
| `test` | Run PHP tests (PHPUnit/Pest) |
| `fmt` | Format code with Laravel Pint |
| `analyse` | Run PHPStan static analysis |
| `packages` | Manage local PHP packages |
| `deploy` | Deploy to Coolify |
| `deploy:status` | Show deployment status |
| `deploy:rollback` | Rollback to previous deployment |
| `deploy:list` | List recent deployments |

## Development Environment

```bash
# Start all services
core php dev
```

Services orchestrated:
- FrankenPHP/Octane (port 8000, HTTPS on 443)
- Vite dev server (port 5173)
- Laravel Horizon (queue workers)
- Laravel Reverb (WebSocket, port 8080)
- Redis (port 6379)

```bash
# View unified logs
core php logs

# Check service status
core php status

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
```

## Building & Serving

```bash
# Build Docker container
core php build

# Build LinuxKit image
core php build --type linuxkit
```

### php serve

Run a production container.

```bash
core php serve [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--name` | Docker image name (required) |
| `--tag` | Image tag (default: latest) |
| `--port` | HTTP port (default: 80) |
| `--https-port` | HTTPS port (default: 443) |
| `-d` | Run in detached mode |
| `--env-file` | Path to environment file |
| `--container` | Container name |

#### Examples

```bash
core php serve --name myapp
core php serve --name myapp -d
core php serve --name myapp --port 8080
```

## Deployment

```bash
# Deploy to Coolify
core php deploy

# Deploy to staging
core php deploy --staging

# Wait for completion
core php deploy --wait

# Check deployment status
core php deploy:status

# List recent deployments
core php deploy:list

# Rollback
core php deploy:rollback
```

## Package Management

Link local packages for development (similar to npm link).

```bash
core php packages <command>
```

| Command | Description |
|---------|-------------|
| `link` | Link local packages by path |
| `unlink` | Unlink packages by name |
| `update` | Update linked packages |
| `list` | List linked packages |

### Examples

```bash
# Link a local package
core php packages link ../my-package

# List linked packages
core php packages list

# Update linked packages
core php packages update

# Unlink
core php packages unlink my-package
```

## SSL/HTTPS

Local SSL with mkcert:

```bash
core php ssl
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

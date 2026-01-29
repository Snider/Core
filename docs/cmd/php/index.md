# core php

Laravel/PHP development tools with FrankenPHP.

## Commands

### Development

| Command | Description |
|---------|-------------|
| [`dev`](#php-dev) | Start development environment |
| [`logs`](#php-logs) | View service logs |
| [`stop`](#php-stop) | Stop all services |
| [`status`](#php-status) | Show service status |
| [`ssl`](#php-ssl) | Setup SSL certificates with mkcert |

### Build & Production

| Command | Description |
|---------|-------------|
| [`build`](#php-build) | Build Docker or LinuxKit image |
| [`serve`](#php-serve) | Run production container |
| [`shell`](#php-shell) | Open shell in running container |

### Code Quality

| Command | Description |
|---------|-------------|
| [`test`](#php-test) | Run PHP tests (PHPUnit/Pest) |
| [`fmt`](#php-fmt) | Format code with Laravel Pint |
| [`analyse`](#php-analyse) | Run PHPStan static analysis |

### Package Management

| Command | Description |
|---------|-------------|
| [`packages link`](#php-packages-link) | Link local packages by path |
| [`packages unlink`](#php-packages-unlink) | Unlink packages by name |
| [`packages update`](#php-packages-update) | Update linked packages |
| [`packages list`](#php-packages-list) | List linked packages |

### Deployment (Coolify)

| Command | Description |
|---------|-------------|
| [`deploy`](#php-deploy) | Deploy to Coolify |
| [`deploy:status`](#php-deploystatus) | Show deployment status |
| [`deploy:rollback`](#php-deployrollback) | Rollback to previous deployment |
| [`deploy:list`](#php-deploylist) | List recent deployments |

---

## php dev

Start the Laravel development environment with all detected services.

```bash
core php dev [flags]
```

### Services Orchestrated

- **FrankenPHP/Octane** - HTTP server (port 8000, HTTPS on 443)
- **Vite** - Frontend dev server (port 5173)
- **Laravel Horizon** - Queue workers
- **Laravel Reverb** - WebSocket server (port 8080)
- **Redis** - Cache and queue backend (port 6379)

### Flags

| Flag | Description |
|------|-------------|
| `--no-vite` | Skip Vite dev server |
| `--no-horizon` | Skip Laravel Horizon |
| `--no-reverb` | Skip Laravel Reverb |
| `--no-redis` | Skip Redis server |
| `--https` | Enable HTTPS with mkcert |
| `--domain` | Domain for SSL certificate (default: from APP_URL) |
| `--port` | FrankenPHP port (default: 8000) |

### Examples

```bash
# Start all detected services
core php dev

# With HTTPS
core php dev --https

# Skip optional services
core php dev --no-horizon --no-reverb
```

---

## php logs

Stream unified logs from all running services.

```bash
core php logs [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--follow` | Follow log output |
| `--service` | Specific service (frankenphp, vite, horizon, reverb, redis) |

---

## php stop

Stop all running Laravel services.

```bash
core php stop
```

---

## php status

Show the status of all Laravel services and project configuration.

```bash
core php status
```

---

## php ssl

Setup local SSL certificates using mkcert.

```bash
core php ssl [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--domain` | Domain for certificate (default: from APP_URL or localhost) |

---

## php build

Build a production-ready container image.

```bash
core php build [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--type` | Build type: `docker` (default) or `linuxkit` |
| `--name` | Image name (default: project directory name) |
| `--tag` | Image tag (default: latest) |
| `--platform` | Target platform (e.g., linux/amd64, linux/arm64) |
| `--dockerfile` | Path to custom Dockerfile |
| `--output` | Output path for LinuxKit image |
| `--format` | LinuxKit format: qcow2 (default), iso, raw, vmdk |
| `--template` | LinuxKit template name (default: server-php) |
| `--no-cache` | Build without cache |

### Examples

```bash
# Build Docker image
core php build

# With custom name and tag
core php build --name myapp --tag v1.0

# Build LinuxKit image
core php build --type linuxkit
```

---

## php serve

Run a production container.

```bash
core php serve [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--name` | Docker image name (required) |
| `--tag` | Image tag (default: latest) |
| `--container` | Container name |
| `--port` | HTTP port (default: 80) |
| `--https-port` | HTTPS port (default: 443) |
| `-d` | Run in detached mode |
| `--env-file` | Path to environment file |

### Examples

```bash
core php serve --name myapp
core php serve --name myapp -d
core php serve --name myapp --port 8080
```

---

## php shell

Open an interactive shell in a running container.

```bash
core php shell <container-id>
```

---

## php test

Run PHP tests using PHPUnit or Pest.

```bash
core php test [flags]
```

Auto-detects Pest if `tests/Pest.php` exists.

### Flags

| Flag | Description |
|------|-------------|
| `--parallel` | Run tests in parallel |
| `--coverage` | Generate code coverage |
| `--filter` | Filter tests by name pattern |
| `--group` | Run only tests in specified group |

### Examples

```bash
core php test
core php test --parallel --coverage
core php test --filter UserTest
```

---

## php fmt

Format PHP code using Laravel Pint.

```bash
core php fmt [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--fix` | Auto-fix formatting issues |
| `--diff` | Show diff of changes |

---

## php analyse

Run PHPStan or Larastan static analysis.

```bash
core php analyse [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--level` | PHPStan analysis level (0-9) |
| `--memory` | Memory limit (e.g., 2G) |

---

## php packages link

Link local PHP packages for development.

```bash
core php packages link <path> [<path>...]
```

Adds path repositories to composer.json with symlink enabled.

---

## php packages unlink

Remove linked packages from composer.json.

```bash
core php packages unlink <name> [<name>...]
```

---

## php packages update

Update linked packages via Composer.

```bash
core php packages update [<name>...]
```

---

## php packages list

List all locally linked packages.

```bash
core php packages list
```

---

## php deploy

Deploy the PHP application to Coolify.

```bash
core php deploy [flags]
```

### Configuration

Requires environment variables in `.env`:
```
COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=your-api-token
COOLIFY_APP_ID=production-app-id
COOLIFY_STAGING_APP_ID=staging-app-id
```

### Flags

| Flag | Description |
|------|-------------|
| `--staging` | Deploy to staging environment |
| `--force` | Force deployment even if no changes detected |
| `--wait` | Wait for deployment to complete |

---

## php deploy:status

Show the status of a deployment.

```bash
core php deploy:status [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--staging` | Check staging environment |
| `--id` | Specific deployment ID |

---

## php deploy:rollback

Rollback to a previous deployment.

```bash
core php deploy:rollback [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--staging` | Rollback staging environment |
| `--id` | Specific deployment ID to rollback to |
| `--wait` | Wait for rollback to complete |

---

## php deploy:list

List recent deployments.

```bash
core php deploy:list [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--staging` | List staging deployments |
| `--limit` | Number of deployments (default: 10) |

---

## Configuration

Optional `.core/php.yaml` - see [Configuration](example.md#configuration) for examples.

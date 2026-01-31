---
name: core-php
description: Use when creating PHP modules, services, or actions in core-* packages.
---

# PHP Framework Patterns

Host UK PHP modules follow strict conventions. Use `core php` commands.

## Module Structure

```
core-{name}/
├── src/
│   ├── Core/              # Namespace: Core\{Name}
│   │   ├── Boot.php       # Module bootstrap (listens to lifecycle events)
│   │   ├── Actions/       # Single-purpose business logic
│   │   └── Models/        # Eloquent models
│   └── Mod/               # Namespace: Core\Mod\{Name} (optional extensions)
├── resources/views/       # Blade templates
├── routes/                # Route definitions
├── database/migrations/   # Migrations
├── tests/                 # Pest tests
└── composer.json
```

## Boot Class Pattern

```php
<?php

declare(strict_types=1);

namespace Core\{Name};

use Core\Php\Events\WebRoutesRegistering;
use Core\Php\Events\AdminPanelBooting;

class Boot
{
    public static array $listens = [
        WebRoutesRegistering::class => 'onWebRoutes',
        AdminPanelBooting::class => ['onAdmin', 10],  // With priority
    ];

    public function onWebRoutes(WebRoutesRegistering $event): void
    {
        $event->router->middleware('web')->group(__DIR__ . '/../routes/web.php');
    }

    public function onAdmin(AdminPanelBooting $event): void
    {
        $event->panel->resources([...]);
    }
}
```

## Action Pattern

```php
<?php

declare(strict_types=1);

namespace Core\{Name}\Actions;

use Core\Php\Action;

class CreateThing
{
    use Action;

    public function handle(User $user, array $data): Thing
    {
        return Thing::create([
            'user_id' => $user->id,
            ...$data,
        ]);
    }
}

// Usage: CreateThing::run($user, $validated);
```

## Multi-Tenant Models

```php
<?php

declare(strict_types=1);

namespace Core\{Name}\Models;

use Core\Tenant\Concerns\BelongsToWorkspace;
use Illuminate\Database\Eloquent\Model;

class Thing extends Model
{
    use BelongsToWorkspace;  // Auto-scopes queries, sets workspace_id

    protected $fillable = ['name', 'workspace_id'];
}
```

## Commands

| Task | Command |
|------|---------|
| Run tests | `core php test` |
| Format | `core php fmt --fix` |
| Analyse | `core php analyse` |
| Dev server | `core php dev` |

## Rules

- Always `declare(strict_types=1);`
- UK English: colour, organisation, centre
- Type hints on all parameters and returns
- Pest for tests, not PHPUnit
- Flux Pro for UI, not vanilla Alpine

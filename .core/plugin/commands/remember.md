---
name: remember
description: Save a fact or decision to context for persistence across compacts
args: <fact to remember>
---

# Remember Context

Save the provided fact to `~/.claude/sessions/context.json`.

## Usage

```
/core:remember Use Action pattern not Service
/core:remember User prefers UK English
/core:remember RFC: minimal state in pre-compact hook
```

## Action

Run this command to save the fact:

```bash
~/.claude/plugins/cache/core/scripts/capture-context.sh "<fact>" "user"
```

Or if running from the plugin directory:

```bash
"${CLAUDE_PLUGIN_ROOT}/scripts/capture-context.sh" "<fact>" "user"
```

The fact will be:
- Stored in context.json (max 20 items)
- Included in pre-compact snapshots
- Auto-cleared after 3 hours of inactivity

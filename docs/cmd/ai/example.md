# AI Examples

## Workflow Example

Complete task management workflow:

```bash
# 1. List available tasks
core ai tasks --status pending

# 2. Auto-select and claim a task
core ai task --auto --claim

# 3. Work on the task...

# 4. Update progress
core ai task:update abc123 --progress 75

# 5. Commit with task reference
core ai task:commit abc123 -m 'implement feature'

# 6. Create PR
core ai task:pr abc123

# 7. Mark complete
core ai task:complete abc123 --output 'Feature implemented and PR created'
```

## Task Filtering

```bash
# By status
core ai tasks --status pending
core ai tasks --status in_progress

# By priority
core ai tasks --priority critical
core ai tasks --priority high

# By labels
core ai tasks --labels bug,urgent

# Combined filters
core ai tasks --status pending --priority high --labels bug
```

## Task Updates

```bash
# Change status
core ai task:update abc123 --status in_progress
core ai task:update abc123 --status blocked

# Update progress
core ai task:update abc123 --progress 25
core ai task:update abc123 --progress 50 --notes 'Halfway done'
core ai task:update abc123 --progress 100
```

## Git Integration

```bash
# Commit with task reference
core ai task:commit abc123 -m 'add authentication'

# With scope
core ai task:commit abc123 -m 'fix login' --scope auth

# Commit and push
core ai task:commit abc123 -m 'complete feature' --push

# Create PR
core ai task:pr abc123

# Draft PR
core ai task:pr abc123 --draft

# PR with labels
core ai task:pr abc123 --labels 'enhancement,ready-for-review'

# PR to different base
core ai task:pr abc123 --base develop
```

## Configuration

### Environment Variables

```env
AGENTIC_TOKEN=your-api-token
AGENTIC_BASE_URL=https://agentic.example.com
```

### ~/.core/agentic.yaml

```yaml
token: your-api-token
base_url: https://agentic.example.com
default_project: my-project
```

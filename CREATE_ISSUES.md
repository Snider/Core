# How to Create GitHub Issues from DX Analysis

This guide helps you quickly create GitHub issues from the DX improvements identified in `DX_IMPROVEMENTS.md`.

## Prerequisites

- GitHub CLI (`gh`) installed and authenticated
- Write access to the repository

## Quick Create (Recommended)

Use this bash script to create individual issues:

```bash
#!/bin/bash
# create-dx-issues.sh

# Issue 1: Missing CONTRIBUTING.md
gh issue create \
  --title "Add CONTRIBUTING.md with contribution guidelines" \
  --label "documentation,good-first-issue,dx-improvement" \
  --body "**Priority:** Critical
**Effort:** ~30 minutes

## Problem
No CONTRIBUTING.md file exists despite having a \"For New Contributors\" section in README. New developers don't know PR conventions, commit message format, or review process.

## Impact
- Contributors waste time on incorrectly formatted PRs
- Maintainers spend time explaining basic workflow
- Higher barriers to entry for new contributors

## Solution
Create \`.github/CONTRIBUTING.md\` with:
- Quick start for contributors
- TDD workflow
- Code standards reference
- Commit message format
- PR process
- Code review expectations

See \`DX_IMPROVEMENTS.md\` for detailed content."

# Issue 2: Missing Development Setup Guide
gh issue create \
  --title "Create comprehensive development setup guide" \
  --label "documentation,dx-improvement,onboarding" \
  --body "**Priority:** Critical
**Effort:** ~45 minutes

## Problem
Developers cloning the repository lack a clear guide on:
- How to install all required dependencies
- How to run CLI vs GUI in development
- How to enable hot-reload
- Tool version requirements

## Impact
- New contributors spend hours on setup
- Inconsistent development environments
- High barrier to first contribution

## Solution
Create \`docs/development-setup.md\` with:
- Prerequisites table (Go, Node.js, Task, Git)
- Initial setup steps
- Development workflows (CLI, GUI, Testing)
- Common tasks reference
- Troubleshooting

See \`DX_IMPROVEMENTS.md\` for full content."

# Issue 3: No CODE_OF_CONDUCT.md
gh issue create \
  --title "Add CODE_OF_CONDUCT.md for community standards" \
  --label "documentation,community,good-first-issue" \
  --body "**Priority:** High
**Effort:** ~15 minutes

## Problem
No code of conduct file exists, making the project less welcoming and unclear about community standards.

## Impact
- Potential contributors may hesitate to engage
- No clear guidelines for handling disputes
- Less professional appearance

## Solution
Add \`.github/CODE_OF_CONDUCT.md\` using Contributor Covenant.

See \`DX_IMPROVEMENTS.md\` for template."

# Issue 4: CLI Help Text Lacks Examples
gh issue create \
  --title "Enhance CLI help text with practical examples" \
  --label "enhancement,cli,dx-improvement" \
  --body "**Priority:** High
**Effort:** ~1 hour

## Problem
When users run \`core <command> --help\`, they see descriptions but no practical examples.

## Impact
- Users struggle to understand command usage
- Increased support requests
- Harder to discover features

## Solution
Add \`Example\` field to all major cobra commands showing common use cases.

Commands to update:
- build, test, dev work, dev health, dev issues
- setup
- All go subcommands
- All php subcommands

See \`DX_IMPROVEMENTS.md\` for implementation details."

# Issue 5: Missing Examples Directory
gh issue create \
  --title "Create examples/ directory with runnable code samples" \
  --label "documentation,examples,dx-improvement" \
  --body "**Priority:** High
**Effort:** ~2 hours

## Problem
No \`examples/\` directory with runnable code samples. Developers must piece together patterns from production code.

## Impact
- Steep learning curve
- Common patterns not discoverable
- More questions in issues/Discord

## Solution
Create \`examples/\` with:
- hello-service/ - Minimal service example
- cli-command/ - Custom CLI command
- error-handling/ - Proper error patterns
- testing-service/ - Service testing patterns
- ipc-actions/ - IPC/Action system usage

Each example needs README, runnable code, and tests.

See \`DX_IMPROVEMENTS.md\` for structure and content."

# Issue 6: Expand Troubleshooting Guide
gh issue create \
  --title "Expand troubleshooting guide with dev-specific issues" \
  --label "documentation,dx-improvement" \
  --body "**Priority:** Medium
**Effort:** ~1 hour

## Problem
\`docs/troubleshooting.md\` missing common developer issues:
- core doctor failure scenarios
- M1 Mac specific issues (CGO, architecture)
- Wails/GUI problems
- Go workspace sync issues

## Impact
- Developers get stuck on known issues
- More support questions
- Time wasted on debugging

## Solution
Expand troubleshooting guide with development environment section.

See \`DX_IMPROVEMENTS.md\` for content."

# Issue 7: Add Task Help Command
gh issue create \
  --title "Add 'task help' command showing common workflows" \
  --label "enhancement,dx-improvement,good-first-issue" \
  --body "**Priority:** Medium
**Effort:** ~30 minutes

## Problem
Taskfile.yml has many tasks but no overview of common workflows.

## Impact
- Developers don't discover useful tasks
- Common workflows not obvious
- More time in documentation

## Solution
Add \`help\` task to Taskfile.yml showing:
- Testing & QA workflows
- Quality assurance tasks
- Building commands
- Code review process
- Multi-repo management
- TDD workflow

See \`DX_IMPROVEMENTS.md\` for implementation."

# Issue 8: Document Environment Variables
gh issue create \
  --title "Document supported environment variables" \
  --label "documentation,dx-improvement" \
  --body "**Priority:** Medium
**Effort:** ~45 minutes

## Problem
No documentation on environment variables the CLI respects. Must read source to discover options.

## Impact
- Configuration options not discoverable
- Developers miss useful features
- Inconsistent behavior

## Solution
Document environment variables in README:
- Logging & Debug (CORE_LOG_LEVEL, etc.)
- Paths & Configuration (CORE_CONFIG_DIR, etc.)
- Build & CI (CORE_BUILD_DIR, GITHUB_TOKEN, etc.)

Include examples for common use cases.

See \`DX_IMPROVEMENTS.md\` for full list."

# Issue 9: Create CLI Output Formatting Guide
gh issue create \
  --title "Create CLI output formatting guide for developers" \
  --label "documentation,dx-improvement" \
  --body "**Priority:** Medium
**Effort:** ~1 hour

## Problem
Core has rich CLI output formatting but no guide for developers on consistent usage.

## Impact
- Inconsistent output across commands
- Developers reinvent formatting code
- Services don't follow UX patterns

## Solution
Create \`docs/cli-output-guide.md\` covering:
- Success/Error messages
- Progress indicators
- Tables and trees
- Color guidelines
- Verbose output patterns
- Error message best practices

See \`DX_IMPROVEMENTS.md\` for content."

# Issue 10: Add Testing Guide
gh issue create \
  --title "Create comprehensive testing guide" \
  --label "documentation,testing,dx-improvement" \
  --body "**Priority:** Low
**Effort:** ~1.5 hours

## Problem
No comprehensive guide on testing patterns:
- Test structure
- Table-driven tests
- Testing services with dependencies
- Mocking patterns
- Testing concurrent code

## Solution
Create \`docs/testing.md\` with complete testing guide.

See \`DX_IMPROVEMENTS.md\` for outline."

# Issue 11: CLI Completion Documentation
gh issue create \
  --title "Document shell completion setup in README" \
  --label "documentation,good-first-issue" \
  --body "**Priority:** Low
**Effort:** ~20 minutes

## Problem
\`core completion\` command exists but no docs on enabling completions in common shells.

## Solution
Add shell completions section to README with examples for:
- Bash
- Zsh
- Fish
- PowerShell

See \`DX_IMPROVEMENTS.md\` for content."

# Issue 12: Linting Configuration Documentation
gh issue create \
  --title "Document linting rules and configuration" \
  --label "documentation,good-first-issue" \
  --body "**Priority:** Low
**Effort:** ~30 minutes

## Problem
\`task lint\` runs golangci-lint but developers don't know what rules apply.

## Solution
- Create/update .golangci.yml with comments
- Document linting rules in README or guide
- Explain how to run locally

See \`DX_IMPROVEMENTS.md\` for details."

# Issue 13: Service Registration Pattern Reconciliation
gh issue create \
  --title "Reconcile service registration pattern documentation" \
  --label "documentation,architecture" \
  --body "**Priority:** Low
**Effort:** ~1 hour

## Problem
README mentions \`core.WithService()\` but code uses \`framework.WithName()\`. Creates confusion about correct pattern.

## Solution
- Audit all docs for service registration examples
- Update to consistent pattern
- Clarify when to use each approach

See \`DX_IMPROVEMENTS.md\` for details."

# Issue 14: Add Architecture Decision Records
gh issue create \
  --title "Create Architecture Decision Records (ADRs)" \
  --label "documentation,architecture,enhancement" \
  --body "**Priority:** Low
**Effort:** ~2 hours

## Problem
No documentation of architectural decisions (why IPC pattern, dual-constructor, etc.).

## Impact
- New developers don't understand design rationale
- Decisions may be questioned/changed without context

## Solution
Create \`docs/adr/\` directory with ADRs:
- Service registration pattern
- IPC bridge vs direct bindings
- Dual constructor pattern
- Others as needed

See \`DX_IMPROVEMENTS.md\` for ADR template."

echo "✅ All 14 DX improvement issues created!"
```

## Manual Creation

If you prefer to create issues manually through GitHub UI:

1. Go to: `https://github.com/host-uk/core/issues/new`
2. Copy title and body from `DX_IMPROVEMENTS.md` for each issue
3. Add appropriate labels from suggestions
4. Submit

## Priority-Based Creation

Create issues in this order for maximum impact:

### Week 1 - Critical Issues
```bash
# Issues 1-2 (CONTRIBUTING.md, Dev Setup Guide)
gh issue create --title "Add CONTRIBUTING.md..." # (see script above)
gh issue create --title "Create comprehensive development setup guide..." 
```

### Week 2 - High Priority
```bash
# Issues 3-5 (CODE_OF_CONDUCT, CLI examples, Examples directory)
gh issue create --title "Add CODE_OF_CONDUCT.md..."
gh issue create --title "Enhance CLI help text..."
gh issue create --title "Create examples/ directory..."
```

### Week 3 - Medium Priority
```bash
# Issues 6-9 (Troubleshooting, Task help, Env vars, CLI formatting)
gh issue create --title "Expand troubleshooting guide..."
gh issue create --title "Add 'task help' command..."
gh issue create --title "Document supported environment variables..."
gh issue create --title "Create CLI output formatting guide..."
```

### Week 4 - Low Priority
```bash
# Issues 10-14 (Testing guide, Completions, Linting, etc.)
gh issue create --title "Create comprehensive testing guide..."
# ... rest as needed
```

## Labels to Use

Make sure these labels exist in your repo:
- `documentation`
- `dx-improvement`
- `good-first-issue`
- `enhancement`
- `cli`
- `examples`
- `testing`
- `architecture`
- `community`
- `onboarding`

Create missing labels:
```bash
gh label create "dx-improvement" --description "Developer Experience improvements" --color "0E8A16"
gh label create "onboarding" --description "New developer onboarding" --color "D4C5F9"
```

## Milestone Creation (Optional)

Group these issues under a milestone:

```bash
gh api repos/host-uk/core/milestones \
  -f title="DX Improvements Q1 2026" \
  -f description="Improve developer experience with better docs and tooling" \
  -f due_on="2026-03-31T00:00:00Z"
```

Then assign issues to the milestone:
```bash
gh issue edit <issue-number> --milestone "DX Improvements Q1 2026"
```

## Project Board (Optional)

Create a project board to track progress:

1. Create board: "DX Improvements"
2. Columns: To Do, In Progress, Review, Done
3. Add all issues to board
4. Track progress visually

## Verification

After creating issues, verify:

```bash
# List all dx-improvement issues
gh issue list --label "dx-improvement"

# Check count
gh issue list --label "dx-improvement" | wc -l
# Should show 14

# View specific issue
gh issue view <issue-number>
```

## Next Steps After Creation

1. **Triage** - Review and adjust priorities based on team feedback
2. **Assign** - Assign issues to team members or mark as `help-wanted`
3. **Link** - Link related issues (e.g., CONTRIBUTING.md needs Dev Setup Guide)
4. **Track** - Monitor progress weekly
5. **Celebrate** - Mark completion and acknowledge contributors

## Quick Reference

| Command | Purpose |
|---------|---------|
| `gh issue create` | Create new issue |
| `gh issue list` | List issues |
| `gh issue view <n>` | View issue details |
| `gh issue edit <n>` | Edit issue |
| `gh issue close <n>` | Close issue |
| `gh label create` | Create label |
| `gh label list` | List labels |

## Support

If you need help:
- Full details in `DX_IMPROVEMENTS.md`
- GitHub CLI docs: https://cli.github.com/manual/
- GitHub Issues docs: https://docs.github.com/en/issues

---

**Note:** This document assumes you have `gh` CLI configured. Install with:
- macOS: `brew install gh`
- Linux: `sudo apt install gh`
- Windows: `winget install GitHub.cli`

Then authenticate: `gh auth login`

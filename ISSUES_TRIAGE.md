# Issues Triage

Generated: 2026-02-02

## Summary

- **Total Open Issues**: 46
- **High Priority**: 6
- **Audit Meta-Issues**: 13 (for Jules AI)
- **Audit Derived Issues**: 20 (created from audits)

---

## High Priority Issues

| # | Title | Labels |
|---|-------|--------|
| 183 | audit: OWASP Top 10 security review | priority:high, jules |
| 189 | audit: Test coverage and quality | priority:high, jules |
| 191 | audit: API design and consistency | priority:high, jules |
| 218 | Increase test coverage for low-coverage packages | priority:high, testing |
| 219 | Add tests for edge cases, error paths, integration | priority:high, testing |
| 168 | feat(crypt): Implement standalone pkg/crypt | priority:high, enhancement |

---

## Audit Meta-Issues (For Jules AI)

These are high-level audit tasks that spawn sub-issues:

| # | Title | Complexity |
|---|-------|------------|
| 183 | audit: OWASP Top 10 security review | large |
| 184 | audit: Authentication and authorization flows | medium |
| 186 | audit: Secrets, credentials, and configuration security | medium |
| 187 | audit: Error handling and logging practices | medium |
| 188 | audit: Code complexity and maintainability | large |
| 189 | audit: Test coverage and quality | large |
| 190 | audit: Performance bottlenecks and optimization | large |
| 191 | audit: API design and consistency | large |
| 192 | audit: Documentation completeness and quality | large |
| 193 | audit: Developer experience (DX) review | large |
| 197 | [Audit] Concurrency and Race Condition Analysis | medium |
| 198 | [Audit] CI/CD Pipeline Security | medium |
| 199 | [Audit] Architecture Patterns | large |
| 201 | [Audit] Error Handling and Recovery | medium |
| 202 | [Audit] Configuration Management | medium |

---

## By Category

### Security (4 issues)

| # | Title | Priority |
|---|-------|----------|
| 221 | Remove StrictHostKeyChecking=no from SSH commands | - |
| 222 | Sanitize user input in execInContainer to prevent injection | - |
| 183 | audit: OWASP Top 10 security review | high |
| 213 | Add logging for security events (authentication, access) | - |

### Testing (3 issues)

| # | Title | Priority |
|---|-------|----------|
| 218 | Increase test coverage for low-coverage packages | high |
| 219 | Add tests for edge cases, error paths, integration | high |
| 220 | Configure branch coverage measurement in test tooling | - |

### Error Handling (4 issues)

| # | Title |
|---|-------|
| 227 | Standardize on cli.Error for user-facing errors, deprecate cli.Fatal |
| 228 | Implement panic recovery mechanism with graceful shutdown |
| 229 | Log all errors at handling point with contextual information |
| 230 | Centralize user-facing error strings in i18n translation files |

### Documentation (6 issues)

| # | Title |
|---|-------|
| 231 | Update README.md to reflect actual configuration management |
| 233 | Add CONTRIBUTING.md with contribution guidelines |
| 234 | Add CHANGELOG.md to track version changes |
| 235 | Add user documentation: user guide, FAQ, troubleshooting |
| 236 | Add configuration documentation to README |
| 237 | Add Architecture Decision Records (ADRs) |

### Architecture (3 issues)

| # | Title |
|---|-------|
| 215 | Refactor Core struct to smaller, focused components |
| 216 | Introduce typed messaging system for IPC (replace interface{}) |
| 232 | Create centralized configuration service |

### Performance (2 issues)

| # | Title |
|---|-------|
| 224 | Add streaming API to pkg/io/local for large file handling |
| 225 | Use background goroutines for long-running operations |

### Logging (3 issues)

| # | Title |
|---|-------|
| 212 | Implement structured logging (JSON format) |
| 213 | Add logging for security events |
| 214 | Implement log retention policy |

### New Features (7 issues)

| # | Title | Priority |
|---|-------|----------|
| 168 | feat(crypt): Implement standalone pkg/crypt | high |
| 167 | feat(config): Implement standalone pkg/config | - |
| 170 | feat(plugin): Consolidate pkg/module into pkg/plugin | - |
| 171 | feat(cli): Implement build variants | - |
| 217 | Implement authentication and authorization features | - |
| 211 | feat(setup): add .core/setup.yaml for dev environment | - |

### Help System (5 issues)

| # | Title | Complexity |
|---|-------|------------|
| 133 | feat(help): Implement display-agnostic help system | large |
| 134 | feat(help): Remove Wails dependencies from pkg/help | large |
| 135 | docs(help): Create help content for core CLI | large |
| 136 | feat(help): Add CLI help command | small |
| 138 | feat(help): Implement Catalog and Topic types | large |
| 139 | feat(help): Implement full-text search | small |

---

## Potential Duplicates / Overlaps

1. **Error Handling**: #187, #201, #227-230 all relate to error handling
2. **Documentation**: #192, #231-237 all relate to documentation
3. **Configuration**: #202, #167, #232 all relate to configuration
4. **Security Audits**: #183, #184, #186, #221, #222 all relate to security

---

## Recommendations

1. **Close audit meta-issues as work is done**: Issues #183-202 are meta-audit issues that should be closed once their derived issues are created/completed.

2. **Link related issues**: Create sub-issue relationships:
   - #187 (audit: error handling) -> #227, #228, #229, #230
   - #192 (audit: docs) -> #231, #233, #234, #235, #236, #237
   - #202 (audit: config) -> #167, #232

3. **Good first issues**: #136, #139 are marked as good first issues

4. **Consider closing duplicates**:
   - #187 vs #201 (both about error handling)
   - #192 vs #231-237 (documentation)

5. **Priority order for development**:
   1. Security fixes (#221, #222)
   2. Test coverage (#218, #219)
   3. Core infrastructure (#168 - crypt, #167 - config)
   4. Error handling standardization (#227-230)
   5. Documentation (#233-237)

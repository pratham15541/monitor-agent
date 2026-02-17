# Version Management

This document explains the unified version source and release process for monitor-tool.

## Single Source of Truth: VERSION file

The **VERSION** file at the repository root (`VERSION`) is the single source of truth for all components.

```
0.1.1
```

## Version Flow

### 1. Development (Automatic)

- Commits to `monitor-agent/**` trigger automatic version bump via `agent-version-bump.yml`
- Uses `scripts/bump_version.sh` to increment patch version (e.g., 0.1.1 → 0.1.2)
- Committed back to main branch

### 2. Building

All build processes read from the VERSION file:

**Local builds:**

```bash
./scripts/build_agent.ps1      # Windows
./scripts/build_agent.sh       # Linux/Mac
```

**CI/CD builds:**

- `agent-build.yml` - reads VERSION file for every commit/PR
- `agent-release.yml` - reads VERSION file for release builds

Version is injected via Go `ldflags`:

```bash
go build -ldflags "-X monitor-agent/cmd.Version=$VERSION"
```

### 3. Releasing

When ready to release:

1. **Manual version bump** (if needed):

   ```bash
   echo "0.2.0" > VERSION
   git add VERSION
   git commit -m "chore: release v0.2.0"
   git push
   ```

2. **Tag the commit:**

   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

3. **CI/CD automatically builds and releases:**
   - `agent-release.yml` triggered on tag push
   - Reads VERSION file (ensures consistency)
   - Builds for all platforms
   - Creates GitHub release with artifacts

## Version Consistency

| Component             | Version Source               |
| --------------------- | ---------------------------- |
| VERSION file          | Manual (source of truth)     |
| Go agent binary       | Injected from VERSION file   |
| Backend (Spring Boot) | Independent (pom.xml)        |
| Frontend (Next.js)    | Independent (package.json)   |
| Git tags              | Manual, should match VERSION |
| CI/CD build           | Reads VERSION file           |
| CI/CD release         | Reads VERSION file           |

## Best Practices

✅ **DO:**

- Keep VERSION file updated before releases
- Use git tags that match VERSION (e.g., v0.1.1)
- Let automatic version bumps handle patch increments
- Use build scripts for local testing

❌ **DON'T:**

- Build directly with `go build` (bypasses version injection)
- Have VERSION file and git tag mismatch
- Manually edit VERSION without understanding the implications
- Create releases with mismatched version numbers

## Example Release Workflow

```bash
# 1. Verify VERSION is correct
cat VERSION  # 0.1.1

# 2. Make final commits
git commit -am "feat: add new feature"
git push

# 3. Tag for release
git tag v0.1.1
git push origin v0.1.1

# 4. GitHub Actions automatically:
#    - Reads VERSION file (0.1.1)
#    - Builds for all platforms
#    - Uploads artifacts to release
```

## Troubleshooting

**Q: Agent shows "dev" version**

- A: Built without `-ldflags` injection. Use `./scripts/build_agent.ps1` or `./scripts/build_agent.sh`

**Q: Release build doesn't match VERSION file**

- A: CI/CD issue. Check `agent-release.yml` is reading VERSION correctly

**Q: Git tag doesn't match VERSION file**

- A: Manual error. Always ensure `git tag -v0.X.X` matches the VERSION file content

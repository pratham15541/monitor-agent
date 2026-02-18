# Version Management

This document explains the unified version source and release process for monitor-tool.

## Single source of truth: VERSION file

The VERSION file at the repository root (VERSION) is the source of truth for agent release builds.

```
0.1.1
```

## Version flow

### 1) Development (automatic)

- Commits to monitor-agent/\*\* trigger automatic version bump via agent-version-bump.yml.
- Uses scripts/bump_version.sh to increment patch version (e.g., 0.1.1 -> 0.1.2).
- Committed back to main branch.

### 2) Building

All agent build processes read from the VERSION file:

Local builds:

```bash
./scripts/build_agent.ps1      # Windows
./scripts/build_agent.sh       # Linux/Mac
```

CI/CD builds:

- agent-build.yml reads VERSION for every commit/PR.
- agent-release.yml reads VERSION for release builds.

Version is injected via Go ldflags:

```bash
go build -ldflags "-X monitor-agent/cmd.Version=$VERSION"
```

### 3) Releasing

When ready to release:

1. Manual version bump (if needed):

```bash
echo "0.2.0" > VERSION
git add VERSION
git commit -m "chore: release v0.2.0"
git push
```

2. Tag the commit:

```bash
git tag v0.2.0
git push origin v0.2.0
```

3. CI/CD automatically builds and releases:

- agent-release.yml is triggered on tag push.
- Reads VERSION file for consistency.
- Builds for all platforms.
- Creates GitHub release with artifacts.

## Version consistency

| Component             | Version source               |
| --------------------- | ---------------------------- |
| VERSION file          | Manual (source of truth)     |
| Go agent binary       | Injected from VERSION file   |
| Backend (Spring Boot) | Independent (pom.xml)        |
| Frontend (Next.js)    | Independent (package.json)   |
| Git tags              | Manual, should match VERSION |
| CI/CD build           | Reads VERSION file           |
| CI/CD release         | Reads VERSION file           |

## Best practices

Do:

- Keep VERSION updated before releases.
- Use git tags that match VERSION (e.g., v0.1.1).
- Let automatic version bumps handle patch increments.
- Use build scripts for local testing.

Do not:

- Build directly with go build (bypasses version injection).
- Allow VERSION and git tag mismatch.
- Create releases with mismatched version numbers.

## Example release workflow

```bash
# 1. Verify VERSION is correct
cat VERSION  # 0.1.1

# 2. Make final commits
git commit -am "feat: add new feature"
git push

# 3. Tag for release
git tag v0.1.1
git push origin v0.1.1
```

## Troubleshooting

Q: Agent shows "dev" version

- Built without ldflags injection. Use ./scripts/build_agent.ps1 or ./scripts/build_agent.sh.

Q: Release build does not match VERSION file

- CI/CD issue. Check agent-release.yml reads VERSION correctly.

Q: Git tag does not match VERSION file

- Manual error. Ensure git tag v0.X.X matches VERSION content.

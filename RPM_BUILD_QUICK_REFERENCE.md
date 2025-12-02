# RPM Build Automation - Quick Reference

## Quick Start - Local Build

### Build RPM locally:
```bash
make rpm VERSION=1.0.0
```

### Build and copy to dist/:
```bash
make rpm-local VERSION=1.0.0
```

### Clean build artifacts:
```bash
make rpm-clean
```

## GitHub Actions

### Automatic Triggers

| Event | Workflow | Action |
|-------|----------|--------|
| Push to any branch | `build.yml` | Builds binaries & RPMs (artifacts not saved) |
| Push tag (e.g., `v1.0.0`) | `release.yml` | Builds all packages and creates release |

### Manual Trigger

From GitHub UI or CLI:
```bash
gh workflow run rpm-build.yml -f version=1.0.0
```

## Installed Files

After `dnf install mdview`:
- `/usr/bin/mdview` - Executable
- `/usr/share/man/man1/mdview.1.gz` - Man page
- License and docs

## Files Created

### Local Files:
- `mdview.spec` - RPM specification file
- `FEDORA_PACKAGING.md` - Detailed Fedora packaging guide
- `RPM_BUILD_AUTOMATION.md` - Complete automation guide
- Updated `Makefile` with RPM targets
- Updated `.github/workflows/release.yml` - Includes RPM builds
- Updated `.github/workflows/build.yml` - Includes RPM builds
- New `.github/workflows/rpm-build.yml` - Manual RPM build workflow

### Build Output:
- `~/rpmbuild/RPMS/x86_64/mdview-*.x86_64.rpm` - Binary RPM
- `~/rpmbuild/SRPMS/mdview-*.src.rpm` - Source RPM
- `dist/mdview-*.rpm` - When using `make rpm-local`

## All Makefile Targets

```bash
make rpm VERSION=X.Y.Z          # Build RPM
make rpm-local VERSION=X.Y.Z    # Build and copy to dist/
make rpm-setup                  # Initialize build environment
make rpm-clean                  # Clean all RPM artifacts
```

## GitHub Release Package

When you push a tag `v1.0.0`:
- Binary distributions for all platforms
- RPM packages (binary and source)
- Debian packages
- All uploaded automatically to GitHub release

## Next Steps

1. Review `RPM_BUILD_AUTOMATION.md` for comprehensive guide
2. Test locally: `make rpm VERSION=1.0.0`
3. Push a tag to trigger GitHub Actions: `git tag v1.0.0 && git push origin v1.0.0`
4. Check GitHub Actions tab for build status
5. Review GitHub Releases page for artifacts

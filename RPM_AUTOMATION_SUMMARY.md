# RPM Package Build Automation - Summary

## What Was Created

I've set up complete automation for building Fedora RPM packages for mdview, both locally and via GitHub Actions.

### Files Created/Modified

#### Configuration Files
1. **`mdview.spec`** - RPM specification file
   - Defines how to build and package mdview
   - Declares dependencies, build steps, and installation rules

#### Documentation
2. **`FEDORA_PACKAGING.md`** - Comprehensive Fedora packaging guide
   - Prerequisites and setup instructions
   - Step-by-step build process
   - Installation and verification
   - Publishing to COPR and Fedora repos

3. **`RPM_BUILD_AUTOMATION.md`** - Complete automation guide
   - Local build instructions
   - Makefile targets reference
   - GitHub Actions workflow details
   - Troubleshooting

4. **`RPM_BUILD_QUICK_REFERENCE.md`** - Quick start guide
   - One-page reference for common tasks
   - Quick commands and file locations

#### Makefile Updates
5. **`Makefile`** - Added RPM targets
   - `make rpm VERSION=X.Y.Z` - Build RPM packages
   - `make rpm-local VERSION=X.Y.Z` - Build and copy to `dist/`
   - `make rpm-setup` - Initialize build environment
   - `make rpm-clean` - Clean all RPM artifacts

#### GitHub Actions Workflows
6. **`.github/workflows/rpm-build.yml`** - NEW
   - Manual RPM-only builds
   - Can be triggered via workflow dispatch
   - Useful for testing or emergency builds

7. **`.github/workflows/release.yml`** - UPDATED
   - Now includes RPM building
   - Builds all platforms and creates GitHub release
   - All artifacts uploaded to release

8. **`.github/workflows/build.yml`** - UPDATED
   - Now includes RPM building on every push
   - Builds binaries and RPMs for validation

## Local Building

### Prerequisites
```bash
sudo dnf install -y rpm-build golang pandoc make git
```

### Build RPM
```bash
# Option 1: Build and keep in ~/rpmbuild/
make rpm VERSION=1.0.0

# Option 2: Build and copy to dist/
make rpm-local VERSION=1.0.0

# Option 3: Clean everything
make rpm-clean
```

### Find Built Packages
```bash
# Binary RPM (x86_64)
~/rpmbuild/RPMS/x86_64/mdview-1.0.0-1.fc*.x86_64.rpm

# Source RPM
~/rpmbuild/SRPMS/mdview-1.0.0-1.fc*.src.rpm

# When using rpm-local:
dist/mdview-*.rpm
```

### Install Locally
```bash
sudo dnf install ~/rpmbuild/RPMS/x86_64/mdview-*.rpm
```

## GitHub Actions Automation

### Build on Every Push
The `build.yml` workflow automatically builds:
- All platform binaries
- RPM packages
- (No artifacts are kept, just validation)

### Release Build on Tag
Push a tag to trigger the `release.yml` workflow:
```bash
git tag v1.0.0
git push origin v1.0.0
```

This will:
1. Build all binaries (Linux amd64/arm64/i386, Windows, macOS, FreeBSD)
2. Build all packages (Debian, RPM)
3. Create a GitHub release
4. Upload all artifacts

### Manual RPM Build
Trigger from GitHub UI:
- Go to Actions → Build RPM Packages → Run workflow
- Enter version number (e.g., `1.0.0`)
- Artifacts are downloaded (not released)

Or via GitHub CLI:
```bash
gh workflow run rpm-build.yml -f version=1.0.0
```

## RPM Package Contents

After installation, mdview provides:
- **Binary**: `/usr/bin/mdview` - The executable
- **Man Page**: `/usr/share/man/man1/mdview.1.gz` - Documentation

Usage:
```bash
mdview --help
mdview --version
man mdview
```

## Version Handling

The version should be passed without the `v` prefix:
```bash
# ✅ Correct
make rpm VERSION=1.0.0

# ❌ Wrong
make rpm VERSION=v1.0.0
```

The Makefile and workflows automatically strip the `v` prefix from git tags.

## RPM Spec File Features

The `mdview.spec` includes:
- **Build requirements**: golang >= 1.21, pandoc
- **Runtime requirements**: xdg-utils (for opening browser)
- **Build process**: Compiles Go binary and generates man page
- **Installation**: Installs binary to `/usr/bin` and man page to `/usr/share/man`
- **Changelog**: Automated changelog generation

## Distributing Packages

### GitHub Releases
All RPMs are automatically uploaded to GitHub releases when you push a tag.

### Fedora COPR
To add mdview to community builds:
1. Create account at https://copr.fedorainfracloud.org/
2. Create a project
3. Link your GitHub repo
4. Enable builds

### Fedora Package Repository
Submit to official Fedora packages:
https://docs.fedoraproject.org/en-US/package-maintainers/

## Key Workflows

```
Local Development → Push Tag → GitHub Actions
                   ↓          ↓
              Build RPM   Release Job
                         ↓
                   GitHub Release
                   (All artifacts)
```

## Troubleshooting Quick Tips

| Issue | Solution |
|-------|----------|
| `rpmbuild: command not found` | `sudo dnf install rpm-build` |
| `go: command not found` | `sudo dnf install golang` |
| `pandoc: command not found` | `sudo dnf install pandoc` |
| Build fails with version errors | Use `VERSION=1.0.0` (no `v` prefix) |
| No RPMs in dist/ | Check `~/rpmbuild/RPMS/x86_64/` |

## Next Steps

1. **Try local build**: `make rpm VERSION=1.0.0`
2. **Review spec file**: `cat mdview.spec`
3. **Read detailed guide**: `cat RPM_BUILD_AUTOMATION.md`
4. **Test with tag**: Create a test tag and push to trigger GitHub Actions
5. **Install and test**: `sudo dnf install ./dist/mdview-*.rpm`

## Documentation Files

Read these for more details:
- `FEDORA_PACKAGING.md` - Complete Fedora packaging guide
- `RPM_BUILD_AUTOMATION.md` - Detailed automation guide
- `RPM_BUILD_QUICK_REFERENCE.md` - Quick command reference

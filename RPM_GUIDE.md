# RPM Package Build Guide

This guide explains how to build Fedora RPM packages for mdview locally and via GitHub Actions.

## Quick Start

### Using the Helper Script (Recommended)

```bash
# Show available commands
./build-rpm.sh help

# Build RPM (uses latest git tag or "dev")
./build-rpm.sh build

# Build specific version
./build-rpm.sh build 1.0.0

# Build and copy to dist/
./build-rpm.sh local 1.0.0

# Setup environment
./build-rpm.sh setup

# Check environment
./build-rpm.sh info

# Clean all artifacts
./build-rpm.sh clean
```

### Using Make Directly

```bash
# Build RPM
make rpm VERSION=1.0.0

# Build and copy to dist/
make rpm-local VERSION=1.0.0

# Initialize environment
make rpm-setup

# Clean artifacts
make rpm-clean
```

## Prerequisites

### On Fedora
```bash
sudo dnf install -y rpm-build golang pandoc make git
```

### Check Setup
```bash
./build-rpm.sh info
```

## Installation

### From Local Build
```bash
# Build
./build-rpm.sh local 1.0.0

# Install
sudo dnf install ./dist/mdview-*.rpm
```

### From RPM File
```bash
sudo dnf install /path/to/mdview-1.0.0-1.fc*.x86_64.rpm
```

## Verify Installation

```bash
which mdview
mdview --version
man mdview
```

## File Organization

```
mdview/
├── mdview.spec                    # RPM specification
├── build-rpm.sh                   # Helper script (this file)
├── Makefile                       # Build automation (rpm targets)
├── FEDORA_PACKAGING.md           # Detailed guide
├── RPM_BUILD_AUTOMATION.md       # Complete automation docs
├── RPM_BUILD_QUICK_REFERENCE.md  # Quick reference
├── RPM_AUTOMATION_SUMMARY.md     # This summary
└── .github/workflows/
    ├── build.yml                  # Build on every push
    ├── release.yml                # Release with RPMs on tag
    └── rpm-build.yml              # Manual RPM builds
```

## GitHub Actions

### Automatic on Push
When you push to any branch, `build.yml` automatically:
- Builds all binaries
- Builds RPM packages
- Validates the build (no artifacts saved)

### Automatic on Tag
When you push a tag (e.g., `v1.0.0`), `release.yml`:
- Builds all binaries and packages
- Creates GitHub release
- Uploads all artifacts (including RPMs)

Example:
```bash
git tag v1.0.0
git push origin v1.0.0
```

### Manual RPM Build
From GitHub UI:
1. Go to Actions tab
2. Select "Build RPM Packages" workflow
3. Click "Run workflow"
4. Enter version (e.g., "1.0.0")
5. Check artifacts

Or via CLI:
```bash
gh workflow run rpm-build.yml -f version=1.0.0
```

## RPM Package Contents

After installation:
- **Binary**: `/usr/bin/mdview`
- **Man page**: `/usr/share/man/man1/mdview.1.gz`
- **License**: `/usr/share/doc/mdview/`

## Build Output Locations

```
~/rpmbuild/RPMS/x86_64/          # Binary RPM (.x86_64.rpm)
~/rpmbuild/SRPMS/                # Source RPM (.src.rpm)
dist/                            # When using rpm-local
```

## Troubleshooting

### Command not found errors
Check the environment:
```bash
./build-rpm.sh info
```

Install missing dependencies:
```bash
sudo dnf install rpm-build golang pandoc make
```

### Build fails
Check the full error:
```bash
make rpm VERSION=1.0.0
```

Common issues:
- Missing dependencies (see above)
- Wrong version format (use `1.0.0` not `v1.0.0`)
- Git not in a repository
- No git history (need at least one commit)

### RPMs not found
Check build directory:
```bash
ls -la ~/rpmbuild/RPMS/x86_64/
ls -la ~/rpmbuild/SRPMS/
```

## Publishing

### To GitHub Releases
Push a tag - GitHub Actions automatically creates release with RPMs.

### To Fedora COPR
1. Create account: https://copr.fedorainfracloud.org/
2. Create project
3. Link GitHub repo
4. Enable builds

### To Fedora Package Repository
Follow: https://docs.fedoraproject.org/en-US/package-maintainers/

## For More Information

- **Local builds**: See `RPM_BUILD_AUTOMATION.md`
- **Quick commands**: See `RPM_BUILD_QUICK_REFERENCE.md`
- **Complete guide**: See `FEDORA_PACKAGING.md`
- **Summary**: See `RPM_AUTOMATION_SUMMARY.md`

## Helper Script Features

The `build-rpm.sh` script provides:
- Dependency checking
- Automatic version detection
- Colored output
- Error handling
- Build environment setup
- Information display

Example output:
```
$ ./build-rpm.sh info
mdview RPM Build Information

✓ Dependencies: OK

Environment:
  RPM Build Dir: /home/user/rpmbuild/
  .rpmmacros: /home/user/.rpmmacros

Version Information:
  Current version: 1.0.0
```

## Development Workflow

1. **Develop locally** - Edit code and test
2. **Build locally** - `./build-rpm.sh build`
3. **Test install** - `sudo dnf install ./dist/mdview-*.rpm`
4. **Push changes** - GitHub validates with `build.yml`
5. **Tag release** - `git tag v1.0.0 && git push origin v1.0.0`
6. **Release created** - GitHub Actions builds and uploads to release

## Support

For issues or questions:
- Check `RPM_BUILD_AUTOMATION.md` for detailed troubleshooting
- Review `.github/workflows/` for automation details
- Examine `mdview.spec` for package configuration

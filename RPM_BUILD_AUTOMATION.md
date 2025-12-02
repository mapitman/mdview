# RPM Build Automation Guide

This document explains how to build RPM packages locally and how GitHub Actions automates the build process.

## Local RPM Building

### Prerequisites

On Fedora:
```bash
sudo dnf install -y rpm-build golang pandoc make git
```

### Quick Build

To build an RPM locally, use the Makefile target:

```bash
# Build for a specific version
make rpm VERSION=1.6.4
```

This will:
1. Set up the RPM build environment in `~/rpmbuild/`
2. Generate the man page from markdown
3. Compile the binary for Linux x86_64
4. Create a source tarball
5. Build the RPM packages (binary and source)
6. Output files to `~/rpmbuild/RPMS/x86_64/` and `~/rpmbuild/SRPMS/`

**Version format:** Use `X.Y.Z` format (e.g., `1.6.4`), not with a `v` prefix.

### Copy Packages Locally

To copy the built RPMs to a `dist/` directory:

```bash
make rpm-local VERSION=1.0.0
```

This creates a `dist/` directory with the RPMs.

### Useful Makefile Targets

The Makefile automates all the RPM build steps:

| Target | Purpose |
|--------|---------|
| `make rpm VERSION=X.Y.Z` | Build RPM packages in `~/rpmbuild/` |
| `make rpm-local VERSION=X.Y.Z` | Build RPM packages and copy them to `dist/` |
| `make rpm-setup` | Initialize the RPM build environment |
| `make rpm-clean` | Remove all RPM build artifacts |

Under the hood, the Makefile:
1. Sets up the RPM build directory structure
2. Generates the man page
3. Compiles the binary
4. Creates a source tarball
5. Runs `rpmbuild` with the spec file
6. Outputs binary and source RPMs

## Installing Locally Built Packages

After building with `make rpm`:

```bash
sudo dnf install ~/rpmbuild/RPMS/x86_64/mdview-*.rpm
```

Or if you used `make rpm-local` (packages in `dist/`):

```bash
sudo dnf install ./dist/mdview-*.rpm
```

For older systems without `dnf`, use `rpm`:

```bash
sudo rpm -ivh ~/rpmbuild/RPMS/x86_64/mdview-*.rpm
```

## Verify Installation

```bash
which mdview
mdview --version
man mdview
```

## GitHub Actions Automation

### Build on Every Push

The `.github/workflows/build.yml` workflow runs on every push and:
- Builds all binary distributions (Linux, Windows, macOS, FreeBSD)
- Builds RPM packages for Fedora
- Does NOT upload artifacts (unless it's a tag)

### Release Build on Tag

When you push a git tag (e.g., `v1.0.0`), the `.github/workflows/release.yml` workflow:

1. **build-binaries job**: Builds all platform binaries and Debian packages on Ubuntu
2. **build-rpm job**: Builds RPM packages in a Fedora container
3. **release job**: Creates a GitHub release with all artifacts

All files are automatically uploaded to the GitHub release.

### Dedicated RPM Workflow

The `.github/workflows/rpm-build.yml` workflow is useful for:
- On-demand RPM builds via workflow dispatch
- Testing RPM builds without full release
- Manual version specification

#### Trigger Manual RPM Build

```bash
# In GitHub Actions UI:
# 1. Go to Actions tab
# 2. Select "Build RPM Packages" workflow
# 3. Click "Run workflow"
# 4. Enter version (e.g., "1.0.0" or "v1.0.0")
```

Or via GitHub CLI:
```bash
gh workflow run rpm-build.yml -f version=1.0.0
```

## Workflow Files

### `.github/workflows/build.yml`
- Runs on every push
- Builds all binaries and RPMs
- No release artifacts

### `.github/workflows/release.yml`
- Runs on tag push
- Builds all binaries and RPMs
- Creates GitHub release with all artifacts

### `.github/workflows/rpm-build.yml`
- Runs on demand or tags
- Builds only RPM packages
- Useful for testing or emergency RPM builds

## RPM Spec File

The `mdview.spec` file contains the RPM package configuration:
- Package metadata (name, version, license, etc.)
- Build dependencies (golang, pandoc)
- Runtime dependencies
- Build process
- Installation instructions

Key variables:
- `%{_version}` - Version (passed via `-D '_version X.Y.Z'`)
- `%{_bindir}` - Binary installation directory (`/usr/bin`)
- `%{_mandir}` - Man page installation directory (`/usr/share/man`)

## Distributing RPMs

### Option 1: GitHub Releases
All RPMs are automatically uploaded to GitHub releases when you push a tag.

### Option 2: Fedora COPR
To add mdview to Fedora's community repository:

1. Create account at https://copr.fedorainfracloud.org/
2. Create a new project
3. Add the GitHub repo as a source
4. Enable builds

See: https://docs.pagure.org/copr.copr/

### Option 3: Fedora Package Repository
Submit to the official Fedora package repository:
- https://docs.fedoraproject.org/en-US/package-maintainers/

## Troubleshooting

### Missing dependencies
If you get "command not found" errors, install the build tools:

```bash
sudo dnf install -y rpm-build golang pandoc make git
```

### Build fails
Check for specific error messages:

```bash
# Verbose output from make
make rpm VERSION=1.6.4 2>&1 | head -50

# Check rpmbuild logs directly
cat ~/rpmbuild/BUILD/mdview-*/build.log
```

### Version format error
Make sure to use the correct version format:

```bash
make rpm VERSION=1.6.4     # ✅ Correct
make rpm VERSION=v1.6.4    # ❌ Wrong
```

### Can't find built packages
Check both locations:

```bash
ls ~/rpmbuild/RPMS/x86_64/mdview-*.rpm       # From 'make rpm'
ls dist/mdview-*.rpm                          # From 'make rpm-local'
```

## RPM Package Contents

After installation via `dnf install`, you get:
- `/usr/bin/mdview` - The mdview executable
- `/usr/share/man/man1/mdview.1.gz` - Compressed man page
- License and documentation files

## Version Format

The VERSION variable should be in format: `X.Y.Z` (e.g., `1.0.0`)
- Don't use the `v` prefix: ❌ `v1.0.0` → ✅ `1.0.0`
- The Makefile and workflows automatically handle the `v` prefix removal

## See Also

- [Fedora Packaging Guide](https://docs.fedoraproject.org/en-US/package-maintainers/)
- [RPM Spec File Reference](https://rpm.org/wiki/RpmSpec)
- [COPR Documentation](https://docs.pagure.org/copr.copr/)

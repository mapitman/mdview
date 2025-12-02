# Building a Fedora RPM Package for mdview

## Overview
This guide explains how to build and create an RPM package for mdview that can be installed on Fedora systems.

## Prerequisites

### On Fedora
Install the necessary build tools:
```bash
sudo dnf install -y rpm-build golang pandoc make
```

### On other systems
You'll need:
- `rpmbuild` or `rpmsign` tools
- Go 1.21+ compiler
- `pandoc` for building the man page
- `make`

## Building the RPM Package

### Option 1: Using Make (Recommended)

The easiest way to build RPM packages is using the Makefile targets:

```bash
# Build RPM and keep in ~/rpmbuild/
make rpm VERSION=1.6.4

# Build RPM and copy to dist/ directory
make rpm-local VERSION=1.6.4

# Initialize RPM build environment (if needed)
make rpm-setup

# Clean all RPM build artifacts
make rpm-clean
```

The `VERSION` should be in format `X.Y.Z` (e.g., `1.6.4`, not `v1.6.4`).

**Built packages location:**
- Binary RPM: `~/rpmbuild/RPMS/x86_64/mdview-*.x86_64.rpm`
- Source RPM: `~/rpmbuild/SRPMS/mdview-*.src.rpm`
- When using `rpm-local`: `dist/mdview-*.rpm`

### Option 2: Using Docker (System-independent)

If you want to build on non-Fedora systems or ensure consistency:

```bash
docker run --rm -v $(pwd):/workspace -w /workspace fedora:latest bash -c '
  dnf install -y rpm-build golang pandoc make git
  mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
  echo "%_topdir %(echo \$HOME)/rpmbuild" > ~/.rpmmacros
  git archive --prefix=mdview-VERSION/ -o ~/rpmbuild/SOURCES/mdview-VERSION.tar.gz HEAD
  cp mdview.spec ~/rpmbuild/SPECS/
  rpmbuild -ba -D "_version VERSION" ~/rpmbuild/SPECS/mdview.spec
  cp ~/rpmbuild/RPMS/x86_64/mdview-*.rpm /workspace/
'
```

## Installing the RPM

Once built, install it on a Fedora system:

```bash
sudo dnf install ./mdview-VERSION-1.fc*.x86_64.rpm
```

Or using older rpm tools:
```bash
sudo rpm -ivh ./mdview-VERSION-1.fc*.x86_64.rpm
```

## Verifying the Installation

```bash
# Check if the binary is installed
which mdview
mdview --version

# Check if the man page is installed
man mdview
```

## Using the spec file with VERSION variable

The spec file uses `%{_version}` as a placeholder. When building, pass the version:

```bash
rpmbuild -ba -D '_version 1.0.0' ~/rpmbuild/SPECS/mdview.spec
```

## Using the Makefile

The Makefile includes integrated RPM building targets:

```bash
# Build RPM
make rpm VERSION=1.6.4

# Build and copy to local dist/ directory
make rpm-local VERSION=1.6.4

# Set up the build environment
make rpm-setup

# Clean all build artifacts
make rpm-clean
```

All the complexity of setting up the RPM build environment, creating source archives, and running rpmbuild is handled automatically.

## Troubleshooting

### Build fails with "pandoc not found"
Install pandoc: `sudo dnf install pandoc`

### Build fails with "go: command not found"
Install Go: `sudo dnf install golang`

### Arch mismatch errors
The current spec file builds for x86_64 (amd64). To build for other architectures:
- Edit the Makefile target in the spec file to use the appropriate architecture
- Or use `setarch` to change the build architecture

## File Locations

After installation, mdview will be located at:
- Binary: `/usr/bin/mdview`
- Man page: `/usr/share/man/man1/mdview.1.gz`

## Publishing to Fedora COPR

To make your package available to others via COPR (Community Projects):

1. Set up an account at https://copr.fedorainfracloud.org/
2. Create a new project
3. Upload the spec file and source code
4. Enable builds for your desired Fedora versions

Refer to COPR documentation: https://docs.pagure.org/copr.copr/

## Notes

- The spec file handles installing both the binary and man page
- The VERSION must be passed at build time
- The package will work on Fedora 36+ (the build requires Go 1.21)
- Consider adding BuildArch declarations if you want to support multiple architectures

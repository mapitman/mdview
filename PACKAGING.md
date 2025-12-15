# Building Packages for mdview

This guide covers building Debian (.deb) and RPM packages for mdview, both locally and via GitHub Actions.

## Prerequisites

### For Debian Packages (Ubuntu/Debian)
```bash
sudo apt-get install -y build-essential golang pandoc fakeroot dpkg-dev
```

### For RPM Packages (Fedora)
```bash
sudo dnf install -y rpm-build golang pandoc make git
```

### Cross-platform with Docker
You can build packages on any system using containers. See "CI Simulation" section below.

## Building Locally

### Debian Packages

Build all Debian packages (amd64, arm64, i386):
```bash
make deb VERSION=1.6.4
```

Find built packages in `dist/`:
- `mdview_1.6.4_amd64.deb`
- `mdview_1.6.4_arm64.deb`
- `mdview_1.6.4_i386.deb`

### RPM Packages

Build RPM packages (x86_64, src.rpm):
```bash
make rpm VERSION=1.6.4
```

Built packages appear in:
- Binary RPM: `~/rpmbuild/RPMS/x86_64/mdview-*.x86_64.rpm`
- Source RPM: `~/rpmbuild/SRPMS/mdview-*.src.rpm`

Alternatively, build and copy to `dist/`:
```bash
make rpm-local VERSION=1.6.4
```

**Note:** Use `X.Y.Z` format (e.g., `1.6.4`), not `vX.Y.Z`.

### Available Makefile Targets

| Target | Purpose |
|--------|---------|
| `make deb VERSION=X.Y.Z` | Build all Debian packages |
| `make rpm VERSION=X.Y.Z` | Build RPM packages in `~/rpmbuild/` |
| `make rpm-local VERSION=X.Y.Z` | Build RPM and copy to `dist/` |
| `make rpm-setup` | Initialize RPM build environment |
| `make rpm-clean` | Clean RPM build artifacts |
| `make ci-sim-ubuntu` | Simulate Ubuntu CI build locally |
| `make ci-sim-fedora` | Simulate Fedora CI build locally |
| `make ci-sim` | Run both CI simulations |

### VCS Information in Binaries

By default, Go embeds version control system (VCS) information into binaries when building from a git repository. This includes the commit hash, modification status, and timestamp.

The Makefile supports a `BUILDVCS_FLAG` variable to control this behavior:
- **Default (local builds)**: VCS info is embedded when available
- **CI/Container builds**: Use `BUILDVCS_FLAG=-buildvcs=false` to disable VCS embedding

**When to disable VCS embedding:**
- Building in CI environments where `.git` directory may not be accessible
- Building in containers/chroots where git metadata is unavailable
- Creating reproducible builds in containerized packaging systems

**Example:**
```bash
# Local build with VCS info (default)
make linux VERSION=1.6.4

# CI/container build without VCS info
BUILDVCS_FLAG=-buildvcs=false make linux VERSION=1.6.4
```

The CI workflows and simulation scripts automatically set this flag for reproducible builds.

## Installing Built Packages

### Debian
```bash
sudo dpkg -i dist/mdview_*.deb
# Or
sudo apt install ./dist/mdview_*.deb
```

### RPM
```bash
sudo dnf install ~/rpmbuild/RPMS/x86_64/mdview-*.rpm
# Or for local dist/ packages
sudo dnf install ./dist/mdview-*.rpm
```

## Verify Installation

```bash
which mdview
mdview --version
man mdview
```

The package installs:
- Binary: `/usr/bin/mdview`
- Man page: `/usr/share/man/man1/mdview.1.gz`

## CI Simulation

You can simulate GitHub Actions builds locally using containers. This is useful for testing changes before pushing.

### Simulate Ubuntu Build
```bash
make ci-sim-ubuntu VERSION=1.6.4
```

This runs the full Ubuntu build process in a container:
- Installs dependencies (including Go 1.21.1)
- Builds binaries
- Creates Debian packages

### Simulate Fedora RPM Build
```bash
make ci-sim-fedora VERSION=1.6.4
```

This runs the Fedora RPM build in a container:
- Installs RPM build tools
- Builds RPM packages
- Outputs to `dist/`

### Simulate Both
```bash
make ci-sim VERSION=1.6.4
```

These commands stream your repository into containers to avoid permission issues with volume mounts.

## GitHub Actions Automation

### Automatic Builds on Release

When you push a git tag (e.g., `v1.6.4`), the release workflow automatically:

1. Builds binaries for all platforms:
   - Linux (amd64, arm64, i386)
   - Windows (amd64, i386)
   - Darwin/macOS (amd64, arm64)
   - FreeBSD (amd64)

2. Builds packages:
   - Debian packages (amd64, arm64, i386)
   - RPM packages (x86_64, src.rpm)

3. Creates a GitHub release with all artifacts

**Trigger a release:**
```bash
git tag v1.6.4
git push origin v1.6.4
```

### Build Workflow

The build workflow runs on every push to validate:
- All platform binaries compile successfully
- No build or test failures

It does **not** build packages or create releases on regular pushes.

## Package Specifications

### Debian Package Structure

Controlled by files in `package/DEBIAN/`:
- `control` - Package metadata, dependencies, description
- Files installed via `package/usr/` directory structure

The Makefile creates architecture-specific packages by copying files and adjusting the control file.

### RPM Package Structure

Controlled by `mdview.spec`:
- Metadata, dependencies, build requirements
- Build steps (compile binary, generate manpage)
- Install steps (copy to /usr/bin and /usr/share/man)

The spec file uses `%{_version}` macro which is passed at build time.

## Troubleshooting

### Debian Build Issues

**Missing fakeroot:**
```bash
sudo apt-get install fakeroot
```

**Architecture not supported:**
Check if your system has the cross-compilation tools:
```bash
sudo apt-get install gcc-aarch64-linux-gnu  # for arm64
sudo apt-get install gcc-arm-linux-gnueabi   # for arm
```

### RPM Build Issues

**pandoc not found:**
```bash
sudo dnf install pandoc
```

**Go version too old:**
mdview requires Go 1.21+. On older Fedora versions:
```bash
sudo dnf install golang
go version  # verify >= 1.21
```

**rpmbuild directory not initialized:**
```bash
make rpm-setup
```

### CI Simulation Issues

**Docker not available:**
Ensure Docker is installed and running:
```bash
docker --version
sudo systemctl start docker
```

**Permission denied:**
Add your user to the docker group:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

**Go version mismatch in Ubuntu container:**
The `ci-sim-ubuntu` script automatically installs Go 1.21.1 if needed.

## Publishing to Repositories

### Fedora COPR

To make RPM packages available via `dnf install mdview`:

1. Create account at https://copr.fedorainfracloud.org/
2. Create a new project
3. Upload spec file and source tarball
4. Enable builds for desired Fedora versions
5. Users can then add your COPR repo and install

### Ubuntu PPA

For Debian packages on Ubuntu:

1. Create Launchpad account
2. Create PPA
3. Upload source package (requires signing with GPG key)
4. Launchpad builds for all Ubuntu versions

### Package Manager Integration

The project is already available via:
- **AUR** (Arch User Repository): `yay -S mdview`
- **deb-get**: `deb-get install mdview`
- **Snap**: `snap install mdview`

## Version Handling

- Git tags should use format `vX.Y.Z` (e.g., `v1.6.4`)
- Makefile VERSION parameter should use `X.Y.Z` (no `v` prefix)
- The build system automatically strips the `v` prefix where needed
- RPM VERSION is sanitized to replace invalid characters with dots

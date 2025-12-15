#!/usr/bin/env bash
set -euo pipefail

# Minimal Fedora-based CI simulation for the RPM build job
# Mount point: repository is mounted at /workdir

VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo 1.6.4)}"
VERSION="${VERSION#v}"
echo "Using VERSION=${VERSION}"

# Install required packages for building RPMs
dnf -y upgrade --refresh || true
dnf install -y --setopt=tsflags=nodocs rpm-build golang make git pandoc rpmdevtools

cd /workdir
export VERSION
export BUILDVCS_FLAG=-buildvcs=false
git config --global --add safe.directory /workdir || true
make rpm-local

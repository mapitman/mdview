#!/usr/bin/env bash
set -euo pipefail

# Minimal Ubuntu-based CI simulation for the 'build' job
# Mount point: repository is mounted at /workdir

VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo 1.6.4)}"
VERSION="${VERSION#v}"
echo "Using VERSION=${VERSION}"

# Ensure required system packages are present
apt-get update -y
apt-get install -y --no-install-recommends build-essential wget tar git pandoc make ca-certificates

# Ensure a compatible Go version (go.mod requires 1.21.1)
GO_VERSION=1.21.1
if command -v go >/dev/null 2>&1; then
	INSTALLED=$(go version | awk '{print $3}' | sed 's/go//')
else
	INSTALLED=0
fi

version_ge() {
	# compare semantic-like versions, returns 0 if $1 >= $2
	printf "%s\n%s\n" "$1" "$2" | sort -V | tail -n1 | grep -qx "$1"
}

if ! version_ge "$INSTALLED" "$GO_VERSION"; then
	echo "Installing Go ${GO_VERSION}"
	TARFILE=/tmp/go${GO_VERSION}.linux-amd64.tar.gz
	CHECKSUM_FILE=/tmp/go${GO_VERSION}.linux-amd64.tar.gz.sha256
	wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O "$TARFILE"
	wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz.sha256" -O "$CHECKSUM_FILE"
	# Verify checksum
	grep -oE '^[0-9a-f]+' "$CHECKSUM_FILE" | awk '{print $1 "  " ENVIRON["TARFILE"]}' > "$CHECKSUM_FILE.checked"
	sha256sum -c "$CHECKSUM_FILE.checked"
	rm -rf /usr/local/go
	tar -C /usr/local -xzf "$TARFILE"
	export PATH=/usr/local/go/bin:$PATH
else
	echo "Found Go ${INSTALLED}, OK"
fi

cd /workdir
export VERSION
export PATH=/usr/local/go/bin:$PATH
make all

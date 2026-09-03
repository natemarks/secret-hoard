#!/usr/bin/env bash
# Remote installer for secret-hoard.
# Usage: curl -fsSL https://raw.githubusercontent.com/natemarks/secret-hoard/main/scripts/install.sh | [sudo] bash -s -- <version> [install-dir]
# Example (system-wide): curl -fsSL ... | sudo bash -s -- v1.2.0
# Example (user): curl -fsSL ... | bash -s -- v1.2.0 "$HOME/bin"
set -euo pipefail

REPO="natemarks/secret-hoard"
GOOS="linux"
GOARCH="amd64"

VERSION="${1:-}"
INSTALL_DIR="${2:-/usr/local/bin}"

if [ -z "$VERSION" ]; then
    echo "Error: version argument required" >&2
    echo "Usage: bash install.sh <version> [install-dir]" >&2
    echo "Example: bash install.sh v1.2.0" >&2
    exit 1
fi

TARBALL="secret-hoard_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Installing secret-hoard ${VERSION} to ${INSTALL_DIR}..."

# Download tarball and checksums
curl -fsSL "${BASE_URL}/${TARBALL}" -o "${TMP_DIR}/${TARBALL}"
curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt"

# Verify checksum (subshell keeps working directory stable)
if ! grep -qF "  ${TARBALL}" "${TMP_DIR}/checksums.txt"; then
    echo "Error: ${TARBALL} not found in checksums.txt" >&2
    exit 1
fi
(cd "${TMP_DIR}" && sha256sum --check --ignore-missing checksums.txt)

# Extract and install
tar -xzf "${TMP_DIR}/${TARBALL}" -C "${TMP_DIR}"
mkdir -p "${INSTALL_DIR}"
for binary in sh-pull sh-push sh-generate sh-contents; do
    if [ -f "${TMP_DIR}/${binary}" ]; then
        install -m 0755 "${TMP_DIR}/${binary}" "${INSTALL_DIR}/${binary}"
        echo "  Installed ${INSTALL_DIR}/${binary}"
    fi
done

echo ""
echo "Installation complete!"
if echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "${INSTALL_DIR} is already in your PATH"
else
    echo "NOTE: Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=${INSTALL_DIR}:\$PATH"
fi

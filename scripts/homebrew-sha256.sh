#!/usr/bin/env bash
# Update packaging/homebrew/reponerve.rb sha256 checksums from GitHub Release assets.
set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "usage: $0 v1.4.0" >&2
  exit 1
fi
VERSION="${VERSION#v}"
FORMULA="packaging/homebrew/reponerve.rb"

declare -A ASSETS=(
  [darwin_arm64]="reponerve_${VERSION}_darwin_arm64.tar.gz"
  [darwin_amd64]="reponerve_${VERSION}_darwin_amd64.tar.gz"
  [linux_arm64]="reponerve_${VERSION}_linux_arm64.tar.gz"
  [linux_amd64]="reponerve_${VERSION}_linux_amd64.tar.gz"
)

for key in "${!ASSETS[@]}"; do
  file="${ASSETS[$key]}"
  url="https://github.com/reponerve/reponerve/releases/download/v${VERSION}/${file}"
  sha=$(curl -fsSL "$url" | shasum -a 256 | awk '{print $1}')
  echo "$key $sha"
done

echo "Update $FORMULA manually or extend this script to sed-replace REPLACE_ON_RELEASE lines."

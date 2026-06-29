#!/usr/bin/env sh
# Install reponerve from GitHub Releases (no Go required).
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/reponerve/reponerve/main/scripts/install.sh | bash
#   REPONERVE_VERSION=v1.3.1 REPONERVE_VERIFY=1 bash scripts/install.sh

set -eu

REPO="${REPONERVE_REPO:-reponerve/reponerve}"
INSTALL_DIR="${REPONERVE_INSTALL_DIR:-$HOME/.local/bin}"
VERIFY="${REPONERVE_VERIFY:-0}"

log() { printf '%s\n' "$*"; }
die() { log "error: $*" >&2; exit 1; }

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

detect_os() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "$os" in
		darwin) echo "darwin" ;;
		linux) echo "linux" ;;
		*) die "unsupported OS: $os (use manual install from docs/install.md)" ;;
	esac
}

detect_arch() {
	arch=$(uname -m)
	case "$arch" in
		x86_64|amd64) echo "amd64" ;;
		aarch64|arm64) echo "arm64" ;;
		*) die "unsupported architecture: $arch" ;;
	esac
}

resolve_version() {
	if [ -n "${REPONERVE_VERSION:-}" ]; then
		printf '%s' "$REPONERVE_VERSION"
		return
	fi
	need_cmd curl
	tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
	[ -n "$tag" ] || die "could not resolve latest release tag; set REPONERVE_VERSION"
	printf '%s' "$tag"
}

version="${REPONERVE_VERSION:-}"
if [ -z "$version" ]; then
	version=$(resolve_version)
fi
case "$version" in v*) ;; *) version="v${version}" ;; esac

os=$(detect_os)
arch=$(detect_arch)
name="reponerve_${version#v}_${os}_${arch}"

case "$os" in
	darwin|linux) archive="${name}.tar.gz" ;;
	*) die "internal error: unknown os $os" ;;
esac

base="https://github.com/${REPO}/releases/download/${version}"
url="${base}/${archive}"

need_cmd curl
need_cmd mkdir
need_cmd mktemp

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT INT HUP TERM

log "Installing RepoNerve ${version} for ${os}/${arch}"
log "Downloading ${url}"

if ! curl -fsSL "$url" -o "${tmpdir}/archive"; then
	die "download failed — set REPONERVE_VERSION to a release with binaries (see docs/install.md)"
fi

if [ "$VERIFY" = "1" ]; then
	if command -v sha256sum >/dev/null 2>&1; then
		checksum_tool="sha256sum"
	else
		need_cmd shasum
		checksum_tool="shasum"
	fi
	checksum_url="${base}/reponerve_${version#v}_checksums.txt"
	curl -fsSL "$checksum_url" -o "${tmpdir}/checksums.txt"
	(
		cd "$tmpdir"
		checksum_line=$(grep " ${archive}\$" checksums.txt || true)
		[ -n "$checksum_line" ] || die "checksum not found for ${archive}"
		set -- $checksum_line
		expected="$1"
		if [ "$checksum_tool" = "sha256sum" ]; then
			printf '%s  archive\n' "$expected" | sha256sum -c -
		else
			actual=$(shasum -a 256 archive | awk '{print $1}')
			[ "$expected" = "$actual" ] || die "checksum mismatch"
		fi
	)
	log "Checksum verified"
fi

case "$archive" in
	*.tar.gz)
		need_cmd tar
		tar -xzf "${tmpdir}/archive" -C "$tmpdir"
		;;
	*.zip)
		need_cmd unzip
		unzip -q "${tmpdir}/archive" -d "$tmpdir"
		;;
esac

[ -f "${tmpdir}/reponerve" ] || die "archive did not contain reponerve binary"

mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
	mv "${tmpdir}/reponerve" "${INSTALL_DIR}/reponerve"
	chmod 755 "${INSTALL_DIR}/reponerve"
else
	need_cmd sudo
	sudo mkdir -p "$INSTALL_DIR"
	sudo mv "${tmpdir}/reponerve" "${INSTALL_DIR}/reponerve"
	sudo chmod 755 "${INSTALL_DIR}/reponerve"
fi

log "Installed to ${INSTALL_DIR}/reponerve"
case ":$PATH:" in
	*":${INSTALL_DIR}:"*) ;;
	*)
		log ""
		log "Add to PATH:"
		log "  export PATH=\"${INSTALL_DIR}:\$PATH\""
		;;
esac

if command -v reponerve >/dev/null 2>&1; then
	log ""
	log "Success — run: reponerve init && reponerve scan"
else
	log ""
	log "Open a new shell or add ${INSTALL_DIR} to PATH, then run: reponerve init"
fi

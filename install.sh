#!/bin/sh

set -eu

REPO="edereagzi/portui"
BINARY="portui"
VERSION="${VERSION:-latest}"
VERIFY_CHECKSUM="${VERIFY_CHECKSUM:-1}"

info() {
    printf '%s\n' "$*"
}

fail() {
    printf 'error: %s\n' "$*" >&2
    exit 1
}

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

download() {
    url="$1"
    out="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --proto '=https' --tlsv1.2 --retry 3 --retry-delay 1 "$url" -o "$out"
        return
    fi
    if command -v wget >/dev/null 2>&1; then
        wget -qO "$out" "$url"
        return
    fi

    fail "curl or wget is required"
}

detect_os() {
    os="$(uname -s)"
    case "$os" in
        Darwin) printf 'darwin\n' ;;
        Linux) printf 'linux\n' ;;
        *) fail "unsupported OS: $os (supported: darwin, linux)" ;;
    esac
}

detect_arch() {
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) printf 'amd64\n' ;;
        arm64|aarch64) printf 'arm64\n' ;;
        *) fail "unsupported architecture: $arch (supported: amd64, arm64)" ;;
    esac
}

resolve_install_dir() {
    if [ -n "${INSTALL_DIR:-}" ]; then
        printf '%s\n' "$INSTALL_DIR"
        return
    fi

    if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
        printf '/usr/local/bin\n'
        return
    fi

    printf '%s/.local/bin\n' "$HOME"
}

resolve_urls() {
    os="$1"
    arch="$2"
    normalized_version="$3"

    asset="${BINARY}_${os}_${arch}.tar.gz"
    if [ "$VERSION" = "latest" ]; then
        base_url="https://github.com/${REPO}/releases/latest/download"
    else
        base_url="https://github.com/${REPO}/releases/download/${normalized_version}"
    fi

    checksum_asset="checksums.txt"
    asset_url="${base_url}/${asset}"
    checksums_url="${base_url}/${checksum_asset}"

    printf '%s\n%s\n%s\n' "$asset" "$asset_url" "$checksums_url"
}

checksum_value() {
    file="$1"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | awk '{print $1}'
        return
    fi
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | awk '{print $1}'
        return
    fi
    fail "checksum verification requested but sha256sum/shasum not found"
}

verify_checksum() {
    archive="$1"
    checksums_file="$2"
    asset="$3"

    expected="$(awk -v target="$asset" '$2 == target { print $1; exit }' "$checksums_file")"
    [ -n "$expected" ] || fail "checksum entry not found for ${asset}"

    actual="$(checksum_value "$archive")"
    [ "$expected" = "$actual" ] || fail "checksum mismatch for ${asset}"
}

install_binary() {
    source_bin="$1"
    dest_dir="$2"
    dest_bin="${dest_dir}/${BINARY}"

    if [ ! -d "$dest_dir" ]; then
        mkdir -p "$dest_dir" 2>/dev/null || fail "cannot create install dir: ${dest_dir}; set INSTALL_DIR to a writable path"
    fi

    if command -v install >/dev/null 2>&1; then
        install -m 0755 "$source_bin" "$dest_bin" 2>/dev/null || fail "cannot write ${dest_bin}; set INSTALL_DIR to a writable path or run with elevated permissions"
    else
        cp "$source_bin" "$dest_bin" 2>/dev/null || fail "cannot write ${dest_bin}; set INSTALL_DIR to a writable path or run with elevated permissions"
        chmod 0755 "$dest_bin"
    fi

    printf '%s\n' "$dest_bin"
}

extract_archive() {
    archive_path="$1"
    tmp_dir="$2"

    need_cmd tar
    tar -xzf "$archive_path" -C "$tmp_dir"
}

main() {
    need_cmd uname
    need_cmd mktemp
    need_cmd awk

    os="$(detect_os)"
    arch="$(detect_arch)"
    install_dir="$(resolve_install_dir)"

    normalized_version="$VERSION"
    if [ "$VERSION" != "latest" ]; then
        normalized_version="v${VERSION#v}"
    fi

    set -- $(resolve_urls "$os" "$arch" "$normalized_version")
    asset="$1"
    asset_url="$2"
    checksums_url="$3"

    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' EXIT INT TERM

    archive_path="${tmp_dir}/${asset}"
    checksums_path="${tmp_dir}/checksums.txt"

    info "Downloading ${asset}..."
    download "$asset_url" "$archive_path"

    if [ "$VERIFY_CHECKSUM" = "1" ]; then
        info "Verifying checksum..."
        download "$checksums_url" "$checksums_path"
        verify_checksum "$archive_path" "$checksums_path" "$asset"
    fi

    info "Extracting archive..."
    extract_archive "$archive_path" "$tmp_dir"
    [ -f "${tmp_dir}/${BINARY}" ] || fail "binary ${BINARY} not found in archive"

    info "Installing to ${install_dir}..."
    installed_path="$(install_binary "${tmp_dir}/${BINARY}" "$install_dir")"

    info "Installed: ${installed_path}"
    "${installed_path}" -v || true

    case ":${PATH}:" in
        *":${install_dir}:"*) ;;
        *)
            info "${install_dir} is not in PATH. Add this line:"
            info "  export PATH=\"${install_dir}:\$PATH\""
            ;;
    esac
}

main "$@"

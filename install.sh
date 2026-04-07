#!/usr/bin/env sh
# install.sh — one-line installer for kt (KarpathyTalk CLI)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yangyang0507/KarpathyTalk-CLI/main/install.sh | sh
#
set -e

REPO="yangyang0507/KarpathyTalk-CLI"
BINARY="kt"
DEFAULT_INSTALL_DIR="/usr/local/bin"

# ── helpers ──────────────────────────────────────────────────────────────────

info()  { printf '  \033[34m•\033[0m %s\n' "$*"; }
ok()    { printf '  \033[32m✓\033[0m %s\n' "$*"; }
err()   { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || err "required tool not found: $1"
}

# ── detect OS ────────────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux"  ;;
    *)      err "Unsupported OS: $(uname -s). Please build from source." ;;
  esac
}

# ── detect architecture ───────────────────────────────────────────────────────

detect_arch() {
  case "$(uname -m)" in
    x86_64)          echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)               err "Unsupported architecture: $(uname -m). Please build from source." ;;
  esac
}

# ── resolve install directory ─────────────────────────────────────────────────

resolve_install_dir() {
  if [ -w "$DEFAULT_INSTALL_DIR" ]; then
    echo "$DEFAULT_INSTALL_DIR"
  elif [ -n "$HOME" ] && [ -d "$HOME/.local/bin" ]; then
    echo "$HOME/.local/bin"
  elif [ -n "$HOME" ]; then
    mkdir -p "$HOME/.local/bin"
    echo "$HOME/.local/bin"
  else
    echo "$DEFAULT_INSTALL_DIR"
  fi
}

# ── fetch latest release tag ──────────────────────────────────────────────────

latest_version() {
  need curl
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)".*/\1/'
}

# ── main ──────────────────────────────────────────────────────────────────────

main() {
  need curl

  OS=$(detect_os)
  ARCH=$(detect_arch)

  info "Detecting latest release…"
  VERSION=$(latest_version)
  [ -n "$VERSION" ] || err "Could not determine latest release. Check your internet connection or visit https://github.com/${REPO}/releases"

  FILENAME="${BINARY}-${OS}-${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

  info "Downloading ${BINARY} ${VERSION} (${OS}/${ARCH})…"
  TMP_DIR=$(mktemp -d)
  TMP_ARCHIVE="${TMP_DIR}/${FILENAME}"
  trap 'rm -rf "$TMP_DIR"' EXIT

  curl -fsSL "$URL" -o "$TMP_ARCHIVE" || err "Download failed. Release asset not found:\n    ${URL}"

  tar -xzf "$TMP_ARCHIVE" -C "$TMP_DIR"
  TMP_BIN="${TMP_DIR}/${BINARY}"
  [ -f "$TMP_BIN" ] || err "Could not find binary '${BINARY}' inside archive"
  chmod +x "$TMP_BIN"

  INSTALL_DIR=$(resolve_install_dir)
  DEST="${INSTALL_DIR}/${BINARY}"

  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_BIN" "$DEST"
  else
    info "Installing to ${INSTALL_DIR} (sudo required)…"
    sudo mv "$TMP_BIN" "$DEST"
  fi

  ok "${BINARY} ${VERSION} installed → ${DEST}"

  # Warn if the install dir is not on PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      printf '\n  \033[33m!\033[0m %s is not on your PATH.\n' "$INSTALL_DIR"
      printf '    Add the following to your shell profile:\n'
      printf '    \033[2mexport PATH="%s:$PATH"\033[0m\n\n' "$INSTALL_DIR"
      ;;
  esac
}

main "$@"

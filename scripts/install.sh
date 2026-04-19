#!/usr/bin/env bash
# ContextFlow installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Luv-Goel/contextflow/main/scripts/install.sh | bash

set -euo pipefail

REPO="Luv-Goel/contextflow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="cf"

detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
    esac
    echo "${OS}-${ARCH}"
}

latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
    echo "🧠 Installing ContextFlow..."
    PLATFORM=$(detect_os)
    VERSION=$(latest_version)
    
    if [[ -z "$VERSION" ]]; then
        echo "Could not determine latest version." >&2
        exit 1
    fi

    URL="https://github.com/${REPO}/releases/download/${VERSION}/cf-${PLATFORM}.tar.gz"
    TMP=$(mktemp -d)
    
    echo "  → Downloading cf ${VERSION} for ${PLATFORM}..."
    curl -fsSL "$URL" | tar xz -C "$TMP"
    
    echo "  → Installing to ${INSTALL_DIR}/cf"
    if [[ -w "$INSTALL_DIR" ]]; then
        mv "${TMP}/cf-${PLATFORM}" "${INSTALL_DIR}/${BINARY}"
        chmod +x "${INSTALL_DIR}/${BINARY}"
    else
        sudo mv "${TMP}/cf-${PLATFORM}" "${INSTALL_DIR}/${BINARY}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY}"
    fi
    
    rm -rf "$TMP"
    
    echo ""
    echo "✅  ContextFlow installed! ($(cf version))"
    echo ""
    echo "Next step — add to your shell config:"
    echo ""
    echo "  bash:  echo 'eval \"\$(cf init --shell bash)\"' >> ~/.bashrc && source ~/.bashrc"
    echo "  zsh:   echo 'eval \"\$(cf init --shell zsh)\"'  >> ~/.zshrc  && source ~/.zshrc"
    echo "  fish:  echo 'cf init --shell fish | source'    >> ~/.config/fish/config.fish"
    echo ""
    echo "Then press Ctrl+R to search, or run 'cf workflows' to explore."
}

main

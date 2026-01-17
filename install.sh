#!/bin/bash
set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CODE_DIR_PATH="restic_helpers"

DATA_DIR="$HOME/.local/share/restic-helpers"
CODE_DIR="$DATA_DIR/$CODE_DIR_PATH"

EXAMPLE_CONFIG_DIR_PATH="examples/config"

CONFIG_DIR="$HOME/.config/restic-helpers"

REPO_BASE="https://raw.githubusercontent.com/catfly/restic-helpers/main"

FILES=(
    pyproject.toml
    core.exclude.txt
    $CODE_DIR_PATH/__init__.py
    $CODE_DIR_PATH/cli.py
    $CODE_DIR_PATH/config.py
    $CODE_DIR_PATH/cron.py
    $CODE_DIR_PATH/launchd.py
    $CODE_DIR_PATH/notify.py
)

EXAMPLE_CONFIG_FILES=(
    $EXAMPLE_CONFIG_DIR_PATH/config.toml
    $EXAMPLE_CONFIG_DIR_PATH/secret.toml
)

echo "Installing restic-helpers to $DATA_DIR"

# Create installation directory
mkdir -p "$DATA_DIR"
rm -rf "$CODE_DIR"
mkdir -p "$CODE_DIR"

# Determine if running from git clone or standalone
if [ -f "$PROJECT_DIR/pyproject.toml" ] && grep -q '^name = "restic-helpers"' "$PROJECT_DIR/pyproject.toml"; then
    echo "Using local files from $PROJECT_DIR"
    for f in "${FILES[@]}"; do
        cp "$PROJECT_DIR/$f" "$DATA_DIR/$f"
    done
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        for f in "${EXAMPLE_CONFIG_FILES[@]}"; do
            cp "$PROJECT_DIR/$f" "$CONFIG_DIR/${f#examples/config/}"
        done
    fi
else
    echo "Downloading files from GitHub..."
    for f in "${FILES[@]}"; do
        curl -fsSL "$REPO_BASE/$f" -o "$DATA_DIR/$f"
    done
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        for f in "${EXAMPLE_CONFIG_FILES[@]}"; do
            curl -fsSL "$REPO_BASE/$f" -o "$CONFIG_DIR/${f#examples/config/}"
        done
    fi
fi

# echo "Removing the venv (if exists) ..."
rm -rf venv

if [ ! -d "venv" ]; then
    echo "Creating venv..."
    python3 -m venv "$DATA_DIR/venv"
fi

"$DATA_DIR/venv/bin/pip" install --upgrade --quiet pip
"$DATA_DIR/venv/bin/pip" install --quiet --editable .

# Create wrapper PROJECT_ROOT
echo "Creating command..."
mkdir -p "$HOME/.local/bin"
cat > "$HOME/.local/bin/restic-helpers" <<'WRAPPER'
#!/bin/bash
exec "$HOME/.local/share/restic-helpers/venv/bin/restic-helpers" "$@"
WRAPPER
chmod +x "$HOME/.local/bin/restic-helpers"

# Check if PATH includes ~/.local/bin
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo ""
    echo "Warning: Add to your ~/.zshrc (or ~/.bashrc):"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo ""
echo "Installation complete!"
echo ""
echo "Reload shell (or add to PATH), then run:"
echo "  restic-helpers --help"

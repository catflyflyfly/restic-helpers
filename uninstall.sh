#!/bin/bash
set -e

echo "Uninstalling restic-helpers..."

# Remove installed files
rm -rf "$HOME/.local/share/restic-helpers"
rm -f "$HOME/.local/bin/restic-helpers"

echo "✅ Uninstalled successfully"
echo ""
echo "Manual cleanup (if needed):"
echo "  - Remove from ~/.zshrc:"
echo "    export PATH=..."
echo "  - Remove repository configs: ~/.config/restic-helpers"

# restic-helpers

A simple python restic wrapper for personal use.

# Features

- **Modern Python CLI** - Clean wrapper around restic commands
- **Single script installation** - No complex setup
- **Config-centric** - Plain text config for restic options, separate secrets file
- **Transparency** - Dry-run and verbose modes show exact commands before execution
- **Monitoring** - Built-in healthchecks.io and Telegram notifications
- **macOS scheduling** - Helpers for launchd integration
- **Reliable** - Configurable exponential backoff retries
- **Safe by default** - Blacklist-centric excludes (better to backup too much than miss critical files)

## Prerequisites

- macOS (for scheduling; backup works on Linux)
- Python 3.7+
- restic installed (`brew install restic`)

## Installation

This installation script is idempotent. You can update safely with this script.

**One-line install (recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/catfly/restic-helpers/main/install.sh | bash
```

**For development or customization:**
```bash
# Clone repository
git clone https://github.com/catfly/restic-helpers.git
cd restic-helpers

# Run installer (copies to ~/.local/share/restic-helpers)
./install.sh
```

### Shell Setup

After installation, add to `~/.zshrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

(Optional) add to `~/.zshenv`:
```bash
[ -f ~/.config/restic-helpers/env.sh ] && source ~/.config/restic-helpers/env.sh
```

Reload shell: `source ~/.zshrc`

## Uninstall
```bash
curl -fsSL https://raw.githubusercontent.com/catfly/restic-helpers/main/uninstall.sh | bash
```

Or manually:
```bash
rm -rf ~/.local/share/restic-helpers
rm -f ~/.local/bin/restic-helpers
# Remove exports from ~/.zshrc
```

## Usage

See `restic-helpers --help` for the command help text.

### Workflow
```bash
# Initialize repo configs
restic-helpers init my_laptop
# Configure your repo parameters.
cd ~/.config/restic-helpers
# Use your repo. This will configure restic ENVs. 
restic-helpers use my_laptop
# Check if your ENVs are correct.
export | grep "^RESTIC"
# Actually initialize your restic repository
# Note: Passwordless ssh is required for SFTP
#   configure ~/.ssh/config, add IdentityFile.
restic init
# Initialize first backup...
# Note: Enable Full Disk Access for terminal app to avoid access errors.
#   1. Go to System Settings, "Privacy & Security" -> "Full Disk Access".
#   2. Click add. Find your terminal app.
restic-helpers backup my_laptop
```

### Schedule Automated Backups
```bash
# Note: Enable Full Disk Access for python to avoid access errors.
source ~/.local/share/restic-helpers/venv/bin/activate
which python | pbcopy
#   1. Go to System Settings, "Privacy & Security" -> "Full Disk Access"
#   2. Click add. Cmd-Shift-G, then paste from clipboard. Click Open.
restic-helpers schedule my_laptop "0 2 * * *" # 02.00 GMT+7
restic-helpers unschedule my_laptop
```

## Development
```bash
# Clone and install in development mode
git clone https://github.com/catfly/restic-helpers.git
cd restic-helpers

# Install
./install.sh

# Test
restic-helpers --help
```

## License

MIT

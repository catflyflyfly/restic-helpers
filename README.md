# restic-helpers

Manage restic backups with scheduling and Telegram notifications.

## Features

- Initialize and manage multiple backup repositories
- Automated scheduling (macOS launchd)
- Telegram notifications (success/failure)
- Retention policy management
- Cross-repository configuration

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

### Initialize Repository
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
restic-helpers backup my_laptop
```

Edit configuration files:
- `~/.config/restic-helpers/repositories/my_laptop/name.txt` - Repository URL
- `~/.config/restic-helpers/repositories/my_laptop/password.txt` - Password
- `~/.config/restic-helpers/repositories/my_laptop/paths.txt` - Paths to backup
- `~/.config/restic-helpers/repositories/my_laptop/exclude.txt` - Custom exclusions

### Configure and Backup
```bash
# Load repository config
restic-helpers configure my_laptop

# Run backup
restic-helpers backup
```

### Schedule Automated Backups
```bash
# Schedule daily at 2:00 AM
restic-helpers schedule my_laptop

# Custom time (14:30)
restic-helpers schedule my_laptop --hour 14 --minute 30

# Remove schedule
restic-helpers unschedule my_laptop
```

Optional: use same token for both if you prefer.

## Commands
```bash
restic-helpers init <repo>              # Initialize repository
restic-helpers configure <repo>         # Load configuration
restic-helpers backup                   # Run backup
restic-helpers schedule <repo> [opts]   # Schedule backup
restic-helpers unschedule <repo>        # Remove schedule
restic-helpers --help                   # Show help
```

## Directory Structure
```
~/.config/restic/repositories/
  └── my_laptop/
      ├── name.txt       # Repository URL
      ├── password.txt   # Repository password
      ├── paths.txt      # Paths to backup
      └── exclude.txt    # Custom exclusions

~/.local/share/restic-helpers/
  ├── venv/            # Python virtual environment
  ├── restic_helpers.py
  ├── requirements.txt
  └── core.exclude.txt

~/.local/bin/
  └── restic-helpers   # Command
```

## Retention Policy

Default retention (configurable in code):
- Daily: 7 snapshots
- Weekly: 4 snapshots
- Monthly: 6 snapshots

## Development
```bash
# Clone and install in development mode
git clone https://github.com/catfly/restic-helpers.git
cd restic-helpers

# Install
./install.sh

# Make changes
vim restic_helpers.py

# Reinstall
./install.sh

# Test
restic-helpers --help
```

## License

MIT

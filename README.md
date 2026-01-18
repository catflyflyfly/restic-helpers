# restic-helpers

A simple restic wrapper for personal use.

## Features

- **Single binary** - No dependencies, just download and run
- **Config-centric** - Plain text config for restic options, separate secrets file
- **Transparency** - Dry-run and verbose modes show exact commands before execution
- **Monitoring** - Built-in healthchecks.io and Telegram notifications
- **macOS scheduling** - Cron-to-launchd conversion for native scheduling
- **Reliable** - Configurable exponential backoff retries
- **Safe by default** - Blacklist-centric excludes (better to backup too much than miss critical files)

## Prerequisites

- macOS (for scheduling; backup works on Linux)
- restic installed (`brew install restic` or equivalent)

## Installation

Download the binary for your platform from [Releases](https://github.com/catfly/restic-helpers/releases).

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/catfly/restic-helpers/releases/latest/download/restic-helpers_darwin_arm64.tar.gz
tar xzf restic-helpers_darwin_arm64.tar.gz

# macOS (Intel)
curl -LO https://github.com/catfly/restic-helpers/releases/latest/download/restic-helpers_darwin_amd64.tar.gz
tar xzf restic-helpers_darwin_amd64.tar.gz

# Linux (x86_64)
curl -LO https://github.com/catfly/restic-helpers/releases/latest/download/restic-helpers_linux_amd64.tar.gz
tar xzf restic-helpers_linux_amd64.tar.gz

# Linux (ARM64)
curl -LO https://github.com/catfly/restic-helpers/releases/latest/download/restic-helpers_linux_arm64.tar.gz
tar xzf restic-helpers_linux_arm64.tar.gz

# Move to PATH
mv restic-helpers /usr/local/bin/
```

Or build from source:
```bash
go install github.com/catfly/restic-helpers/cmd/restic-helpers@latest
```

## Usage

See `restic-helpers --help` for all commands.

### Workflow

```bash
# Initialize repo config (creates ~/.config/restic-helpers/ and repo config files)
restic-helpers init my_laptop

# Edit your repo config
cd ~/.config/restic-helpers/repositories/my_laptop
# - name.txt: repository URL/path
# - password.txt: repository password
# - paths.txt: paths to backup
# - exclude.txt: additional exclusion patterns
# - healthcheck.txt: healthchecks.io URL (optional)

# Set restic environment variables for this repo
restic-helpers use my_laptop

# Verify environment
export | grep "^RESTIC"

# Initialize the restic repository (first time only)
# Note: For SFTP, configure passwordless SSH in ~/.ssh/config
restic init

# Run a backup
# Note: On macOS, enable Full Disk Access for terminal app
#   System Settings -> Privacy & Security -> Full Disk Access
restic-helpers backup my_laptop
```

### Schedule Automated Backups (macOS)

```bash
# Schedule daily backup at 2am
restic-helpers schedule my_laptop "0 2 * * *"

# Remove schedule
restic-helpers unschedule my_laptop
```

Note: For scheduled backups, enable Full Disk Access for the binary:
1. System Settings -> Privacy & Security -> Full Disk Access
2. Add `/usr/local/bin/restic-helpers`

## Configuration

Config files are stored in `~/.config/restic-helpers/`:

```
~/.config/restic-helpers/
├── config.toml          # Global settings (prune retention, retry, etc.)
├── secret.toml          # Sensitive values (Telegram bot token)
├── core.exclude.txt     # Common exclusion patterns
└── repositories/
    └── my_laptop/
        ├── name.txt
        ├── password.txt
        ├── paths.txt
        ├── exclude.txt
        └── healthcheck.txt
```

## License

MIT

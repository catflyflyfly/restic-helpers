# restic-helpers

Personal shell functions and documentation for managing [restic](https://restic.net/) backups on macOS.

This repository serves as both a toolkit and a reference document - a way to automate backups while maintaining clear documentation of what's happening and why. Built primarily for personal use on macOS systems.

## Prerequisites

- **restic** - Backup tool. Install via:

  ```bash
    brew install restic
  ```

## Installation

Run this in terminal to update your `.zshrc`:

```bash
echo 'export RESTIC_HELPERS_DIR="$HOME/restic-helpers"' >> ~/.zshrc
echo '[ -f "$RESTIC_HELPERS_DIR/bin/functions.sh" ] && source "$RESTIC_HELPERS_DIR/bin/functions.sh"' >> ~/.zshrc
source ~/.zshrc
```

## Configuration

### 1. Initialize Repository Configuration

Create a new repository configuration:

```bash
init_restic_repo REPO_NAME
```

Example:

```bash
init_restic_repo my_macbook
```

This creates a directory at `repos/REPO_NAME/` with template configuration files.

### 2. Edit Configuration Files

Fill in the generated files:

- `name.txt` - Restic repository URL (e.g., `sftp:user@host:/path/to/backup`)
- `password.txt` - Repository encryption password
- `exclude.txt` - Repository-specific exclusion patterns
- `paths.txt` - Paths to backup (one per line)

For repository URL formats, see [Restic documentation](https://restic.readthedocs.io/en/stable/030_preparing_a_new_repo.html).

### 3. Initialize Restic Repository

Configure and initialize the restic repository:

```bash
configure_restic_repo REPO_NAME
restic init \
  --repository-file="$X_RESTIC_REPOSITORY_FILE" \
  --password-file="$X_RESTIC_PASSWORD_FILE"
```

## Usage

### Backup

Run a backup:

```bash
configure_restic_repo REPO_NAME
restic_backup
```

### Export Environment Variables

For easier command-line usage:

```bash
configure_restic_repo REPO_NAME
export_restic_env
restic snapshots  # Now uses standard restic env vars
```

For more restic commands, see [Restic documentation](https://restic.readthedocs.io/en/stable/).

## Automation

### macOS (launchd)

Schedule automated backups:

```bash
schedule_restic_backup_macos REPO_NAME [HOUR] [MINUTE]
```

Example - run daily at 2:00 AM:

```bash
schedule_restic_backup_macos my_macbook
```

Example - run daily at 2:30 PM:

```bash
schedule_restic_backup_macos my_macbook 14 30
```

### Unschedule

Remove scheduled backup:

```bash
unschedule_restic_backup_macos REPO_NAME
```

### Notes

- Scheduled tasks only run when the Mac is awake and powered on
- For reliable backups, keep your Mac plugged in during scheduled times
- Logs are stored in `/tmp/restic-backup-REPO_NAME/`
- For repository integrity checks, schedule them on your NAS instead (always-on, faster)

## License

MIT License - see [LICENSE](LICENSE) file for details.

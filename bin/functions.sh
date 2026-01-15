#!/bin/bash

# Initialize a new restic repository configuration.
# Creates directory structure and template files for a new repository.
#
# Usage: init_restic_repo REPO
# Arguments:
#   REPO - Repository nickname (e.g., my_macbook)
# Examples:
#   init_restic_repo my_macbook
#   init_restic_repo work_laptop
init_restic_repo() {
  if [ -z "$1" ]; then
    echo "Error: REPO name required"
    echo "Usage: init_restic_repo REPO"
    return 1
  fi
  
  local REPO="$1"
  local RESTIC_REPOS_DIR="$HOME/.config/restic/repositories"
  local REPO_DIR="$RESTIC_REPOS_DIR/$REPO"
  
  if [ -d "$REPO_DIR" ]; then
    echo "Error: Repository '$REPO' already exists"
    return 1
  fi
  
  mkdir -p "$REPO_DIR"
  
  touch "$REPO_DIR/name.txt"
  touch "$REPO_DIR/password.txt"
  touch "$REPO_DIR/exclude.txt"
  touch "$REPO_DIR/paths.txt"
  
  chmod 600 "$REPO_DIR/password.txt"
  
  echo "Created repository configuration: $REPO"
  echo "Edit the following files:"
  echo "  $REPO_DIR/name.txt       - Repository URL"
  echo "  $REPO_DIR/password.txt   - Repository password"
  echo "  $REPO_DIR/exclude.txt    - Custom exclusions"
  echo "  $REPO_DIR/paths.txt      - Paths to backup"
}

# Load restic repository configuration by setting environment variables
# for repository path, password, exclusions, and backup paths.
#
# Usage: configure_restic_repo REPO
# Arguments:
#   REPO - Repository name (matches directory in repos/)
# Examples:
#   configure_restic_repo catbook
#   configure_restic_repo work
configure_restic_repo() {
  if [ -z "$RESTIC_HELPERS_DIR" ]; then
    echo "Error: RESTIC_HELPERS_DIR not set"
    return 1
  fi
  
  if [ -z "$1" ]; then
    echo "Error: REPO name required"
    echo "Usage: configure_restic_repo REPO"
    return 1
  fi
  
  local REPO="$1"
  local RESTIC_REPOS_DIR="$HOME/.config/restic/repositories"
  local REPO_DIR="$RESTIC_REPOS_DIR/$REPO"
  local CORE_REPO_DIR="$RESTIC_REPOS_DIR/core"
  
  export RESTIC_REPOSITORY_FILE="$REPO_DIR/name.txt"
  export RESTIC_PASSWORD_FILE="$REPO_DIR/password.txt"
  export X_RESTIC_CORE_EXCLUDE_FILE="$RESTIC_HELPERS_DIR/core.exclude.txt"
  export X_RESTIC_EXCLUDE_FILE="$REPO_DIR/exclude.txt"
  export X_RESTIC_PATHS_FILE="$REPO_DIR/paths.txt"
  
  echo "Configured restic repository: $(cat "$RESTIC_REPOSITORY_FILE")"
}

# Perform restic backup and prune old snapshots according to retention policy.
#
# Usage: restic_backup
# Prerequisites: Must call configure_restic_repo first
# Examples:
#   configure_restic_repo catbook
#   restic_backup
restic_backup() {
  if [ -z "$RESTIC_REPOSITORY_FILE" ]; then
    echo "Error: No repository configured. Call configure_restic_repo first."
    return 1
  fi

  # Run backup - ignore stdout, stderr to temp file
  restic backup \
    --repository-file="$RESTIC_REPOSITORY_FILE" \
    --password-file="$RESTIC_PASSWORD_FILE" \
    --files-from="$X_RESTIC_PATHS_FILE" \
    --exclude-file="$X_RESTIC_CORE_EXCLUDE_FILE" \
    --exclude-file="$X_RESTIC_EXCLUDE_FILE" \
    --exclude-caches \
    2> >(tee /tmp/restic-backup-error.txt >&2)

  EXIT_CODE=$?

  # Send notification
  if [ $EXIT_CODE -eq 0 ]; then
    BOT_TOKEN="$TELEGRAM_INFO_BOT_TOKEN"
    MSG="✅ Backup succeeded"
  else
    BOT_TOKEN="$TELEGRAM_ERROR_BOT_TOKEN"
    ERROR=$(cat /tmp/restic-backup-error.txt)
    MSG="❌ Backup failed\n$ERROR"
  fi

  # Skip if telegram not configured
  if [ -n "$TELEGRAM_CHAT_ID" ] && [ -n "$BOT_TOKEN" ]; then
    jq -n \
      --arg chat_id "$TELEGRAM_CHAT_ID" \
      --arg msg "$MSG" \
      '{chat_id: $chat_id, text: $msg}' | \
    curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
      -H "Content-Type: application/json" \
      -d @-
  fi

  # Exit only if failed; do not proceed to prune
  if [ $EXIT_CODE -ne 0 ]; then
    return $EXIT_CODE
  fi
  
  restic forget \
    --repository-file="$RESTIC_REPOSITORY_FILE" \
    --password-file="$RESTIC_PASSWORD_FILE" \
    --keep-daily 7 \
    --keep-weekly 4 \
    --keep-monthly 6 \
    --prune
}

# Schedule automated restic backup using launchd.
# Generates plist file and loads it into launchd.
#
# Usage: schedule_restic_backup_macos REPO_NAME [HOUR] [MINUTE]
# Arguments:
#   REPO_NAME - Repository name (e.g., catbook_catfly)
#   HOUR      - Hour to run (0-23, default: 2)
#   MINUTE    - Minute to run (0-59, default: 0)
# Examples:
#   schedule_restic_backup_macos catbook_catfly
#   schedule_restic_backup_macos catbook_catfly 14 30
schedule_restic_backup_macos() {
  if [ -z "$RESTIC_HELPERS_DIR" ]; then
    echo "Error: RESTIC_HELPERS_DIR not set"
    return 1
  fi
  
  if [ -z "$1" ]; then
    echo "Error: REPO_NAME required"
    echo "Usage: schedule_restic_backup_macos REPO_NAME [HOUR] [MINUTE]"
    return 1
  fi

  local REPO_NAME="$1"
  local HOUR="${2:-2}"
  local MINUTE="${3:-0}"
  local LABEL="local.restic.backup.$REPO_NAME"
  local PLIST_FILE="$HOME/Library/LaunchAgents/$LABEL.plist"
  local LOG_DIR="/tmp/restic-backup-$REPO_NAME"
  local LAUNCHER="$RESTIC_HELPERS_DIR/bin/restic-backup-launcher"

  mkdir -p "$LOG_DIR"
  mkdir -p "$HOME/Library/LaunchAgents"

  cat > "$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$LABEL</string>
    <key>ProgramArguments</key>
    <array>
        <string>$LAUNCHER</string>
        <string>$REPO_NAME</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>$HOUR</integer>
        <key>Minute</key>
        <integer>$MINUTE</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/output.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/error.log</string>
</dict>
</plist>
EOF

  echo "Created: $PLIST_FILE"
  
  launchctl load "$PLIST_FILE"
  echo "Loaded: $LABEL"
  echo "Scheduled to run daily at $HOUR:$(printf '%02d' $MINUTE)"
}

# Uninstall scheduled restic backup from launchd.
# Unloads plist and removes it from LaunchAgents.
#
# Usage: unschedule_restic_backup_macos REPO_NAME
# Arguments:
#   REPO_NAME - Repository name (e.g., catbook_catfly)
# Examples:
#   unschedule_restic_backup_macos catbook_catfly
unschedule_restic_backup_macos() {
  if [ -z "$1" ]; then
    echo "Error: REPO_NAME required"
    echo "Usage: unschedule_restic_backup_macos REPO_NAME"
    return 1
  fi

  local REPO_NAME="$1"
  local LABEL="local.restic.backup.$REPO_NAME"
  local PLIST_FILE="$HOME/Library/LaunchAgents/$LABEL.plist"

  if [ ! -f "$PLIST_FILE" ]; then
    echo "Error: Schedule not found for $REPO_NAME"
    return 1
  fi

  launchctl unload "$PLIST_FILE"
  rm "$PLIST_FILE"
  
  echo "Unloaded and removed: $LABEL"
}

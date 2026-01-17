"""
Restic backup helper - Manage restic backups with scheduling and notifications
"""

import os
from pathlib import Path

PROG_NAME = "restic-helpers"
__version__ = "0.1.0"

CONFIG_DIR = Path.home() / ".config" / PROG_NAME
CONFIG_REPO_DIR = CONFIG_DIR / "repositories"
DATA_DIR = Path(os.environ.get("DATA_DIR", Path.home() / ".local/share" / PROG_NAME))

"""Launchd plist generation for macOS scheduling"""

import plistlib
from pathlib import Path

import click
from croniter import croniter, CroniterBadCronError

from .cron import cron_to_launchd, InvalidCronExpression


def generate_launchd_plist(repo_name: str, cron_expr: str):
    """Generate launchd plist data for scheduled backup"""
    restic_helpers_bin = Path.home() / ".local/bin/restic-helpers"
    log_dir = Path.home() / ".local/state/restic-helpers" / repo_name

    try:
        cron = croniter(cron_expr)
    except CroniterBadCronError as e:
        raise InvalidCronExpression(
            f"Invalid cron expression '{cron_expr}': {e}"
        ) from None
    schedule = cron_to_launchd(cron)

    return {
        "Label": label(repo_name),
        "ProgramArguments": [str(restic_helpers_bin), "backup", repo_name],
        "StartCalendarInterval": schedule,
        "EnvironmentVariables": {
            "PATH": "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin",
        },
        "StandardOutPath": str(log_dir / "output.log"),
        "StandardErrorPath": str(log_dir / "error.log"),
    }


def write_launchd_plist(ctx, repo_name, cron_expr):
    """Write plist file for scheduled backup"""
    plist_file = Path.home() / "Library/LaunchAgents" / f"{label(repo_name)}.plist"
    log_dir = Path.home() / ".local/state/restic-helpers" / repo_name

    plist_data = generate_launchd_plist(repo_name, cron_expr)

    if ctx.dry_run:
        click.echo(f"# Would write: {plist_file}")
        click.echo(plistlib.dumps(plist_data).decode())
        return plist_file

    plist_file.parent.mkdir(parents=True, exist_ok=True)
    log_dir.mkdir(parents=True, exist_ok=True)

    with open(plist_file, "wb") as f:
        plistlib.dump(plist_data, f)

    return plist_file


def label(repo_name):
    return f"local.restic.backup.{repo_name}"

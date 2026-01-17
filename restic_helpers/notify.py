"""Telegram notification support"""

import json
from enum import Enum

import click
import requests

from . import CONFIG_REPO_DIR
from .config import Context


def send_error_to_telegram(ctx: Context, operation: str, repo_name: str, stderr: str):
    """Send error notification to Telegram"""
    message = f"❌ {operation} failed for {repo_name}\n{stderr}"
    telegram = ctx.config.telegram

    if not telegram.enabled:
        if ctx.verbose:
            click.echo("Telegram notifications disabled")
        return

    if ctx.verbose:
        click.echo("Preparing Telegram notification")

    chat_id = telegram.chat_id
    bot_token = telegram.bot_token

    if ctx.verbose:
        click.echo(f"  chat_id: {'set' if chat_id else 'missing'}")
        click.echo(f"  bot_token: {'set' if bot_token else 'missing'}")

    missing = []
    if not chat_id:
        missing.append("chat_id")
    if not bot_token:
        missing.append("bot_token")

    if missing:
        if ctx.verbose:
            click.echo(f"  Skipping: missing {', '.join(missing)}")
        return

    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
    data = {"chat_id": chat_id, "text": message}

    if ctx.verbose:
        click.echo(f"  URL: {url}")
        click.echo(f"  Payload: {data}")

    if ctx.dry_run:
        click.echo(
            f"""curl -X POST "{url}" \\
  -H "Content-Type: application/json" \\
  -d '{json.dumps(data)}'
"""
        )
    else:
        if ctx.verbose:
            click.echo("  Sending request...")
        try:
            requests.post(url, json=data, timeout=10)
            if ctx.verbose:
                click.echo("  Sent successfully")
        except Exception as e:
            click.echo(f"Warning: Failed to send Telegram notification: {e}", err=True)


class HealthcheckStatus(Enum):
    START = "/start"
    SUCCESS = ""
    FAIL = "/fail"


def ping_healthcheck(ctx: Context, repo_name: str, status: HealthcheckStatus):
    """Ping healthchecks.io"""
    hc_file = CONFIG_REPO_DIR / repo_name / "healthcheck.txt"

    if not hc_file.exists():
        if ctx.verbose:
            click.echo(f"  No healthcheck configured for {repo_name}")
        return

    url = hc_file.read_text().strip()
    if not url:
        return

    ping_url = f"{url.rstrip('/')}{status.value}"

    if ctx.dry_run:
        click.echo(f"# Would ping: {ping_url}")
        return

    if ctx.verbose:
        click.echo(f"  Pinging healthcheck: {ping_url}")

    try:
        requests.get(ping_url, timeout=10)
    except Exception as e:
        click.echo(f"Warning: Failed to ping healthcheck: {e}", err=True)

"""Configuration models and loading"""

import os

import click
import tomllib
from pydantic import BaseModel, Field
from tenacity import (
    Retrying,
    retry_if_exception_type,
    stop_after_attempt,
    wait_exponential,
)

from . import CONFIG_DIR, CONFIG_REPO_DIR


def setattr_from_env(obj, env_prefix, key, env=os.environ):
    """Set attribute from env var, auto-generating name and inferring type"""
    env_key = f"{env_prefix}{key.upper()}"

    if value := env.get(env_key):
        # Infer type from existing attribute
        current = getattr(obj, key)
        if isinstance(current, bool):
            setattr(obj, key, value.lower() in ("true", "1", "yes"))
        elif isinstance(current, int):
            setattr(obj, key, int(value))
        elif isinstance(current, float):
            setattr(obj, key, float(value))
        else:
            setattr(obj, key, value)


class TelegramConfig(BaseModel):
    enabled: bool = True
    chat_id: str | None = None
    bot_token: str | None = None

    def model_post_init(self, __context):
        self.override_from_env()

    def override_from_env(self):
        prefix = "X_RESTIC_TELEGRAM_"
        setattr_from_env(self, prefix, "enabled")
        setattr_from_env(self, prefix, "chat_id")
        setattr_from_env(self, prefix, "bot_token")


class PruneConfig(BaseModel):
    keep_daily: int = 7
    keep_weekly: int = 4
    keep_monthly: int = 6

    def model_post_init(self, __context):
        self.override_from_env()

    def override_from_env(self):
        prefix = "X_RESTIC_PRUNE_"
        setattr_from_env(self, prefix, "keep_daily")
        setattr_from_env(self, prefix, "keep_weekly")
        setattr_from_env(self, prefix, "keep_monthly")


class RetryConfig(BaseModel):
    multiplier: int = 1
    max_attempts: int = 5
    backoff_min: int = 1
    backoff_max: int = 60
    exp_base: int = 2

    def model_post_init(self, __context):
        self.override_from_env()

    def override_from_env(self):
        prefix = "X_RESTIC_RETRY_"
        setattr_from_env(self, prefix, "multiplier")
        setattr_from_env(self, prefix, "max_attempts")
        setattr_from_env(self, prefix, "backoff_min")
        setattr_from_env(self, prefix, "backoff_max")
        setattr_from_env(self, prefix, "exp_base")

    def retryer(self, retry_on: retry_if_exception_type, verbose: bool):
        kwargs = {
            "stop": stop_after_attempt(self.max_attempts),
            "wait": wait_exponential(
                multiplier=self.multiplier,
                min=self.backoff_min,
                max=self.backoff_max,
                exp_base=self.exp_base,
            ),
            "retry": retry_on or retry_if_exception_type(Exception),
            "reraise": True,
        }

        if verbose:
            import click

            kwargs["before"] = lambda info: click.echo(
                f"Attempt {info.attempt_number}..."
            )
            kwargs["before_sleep"] = lambda info: click.echo(
                f"Retrying in {info.idle_for:.1f}s..."
            )

        return Retrying(**kwargs)


class RepoConfig(BaseModel):
    prune: PruneConfig = PruneConfig()


class Config(BaseModel):
    telegram: TelegramConfig = Field(default_factory=TelegramConfig)
    prune: PruneConfig = Field(default_factory=PruneConfig)
    retry: RetryConfig = Field(default_factory=RetryConfig)
    repos: dict[str, RepoConfig] = {}


def load_config() -> Config:
    """Load global config and secret"""
    raw = {}

    config_file = CONFIG_DIR / "config.toml"
    if config_file.exists():
        with open(config_file, "rb") as f:
            raw = tomllib.load(f)

    secret_file = CONFIG_DIR / "secret.toml"
    if secret_file.exists():
        with open(secret_file, "rb") as f:
            secret = tomllib.load(f)
            for key, value in secret.items():
                if (
                    key in raw
                    and isinstance(raw[key], dict)
                    and isinstance(value, dict)
                ):
                    raw[key].update(value)
                else:
                    raw[key] = value

    config = Config(**raw)

    # Load per-repo configs
    if CONFIG_REPO_DIR.exists():
        for repo_dir in CONFIG_REPO_DIR.iterdir():
            if not repo_dir.is_dir():
                continue

            repo_raw = {}
            prune_file = repo_dir / "prune.toml"
            if prune_file.exists():
                with open(prune_file, "rb") as f:
                    repo_raw["prune"] = tomllib.load(f)

            # Merge with global defaults
            repo_config = RepoConfig(
                prune=PruneConfig(
                    **{**config.prune.model_dump(), **repo_raw.get("prune", {})}
                )
            )
            config.repos[repo_dir.name] = repo_config

    return config


class Context:
    def __init__(self, dry_run=False, verbose=False):
        self.dry_run = dry_run
        self.verbose = verbose
        self.config = load_config()
        if verbose:
            click.echo("Config loaded:")
            click.echo(self.config.model_dump_json(indent=2))

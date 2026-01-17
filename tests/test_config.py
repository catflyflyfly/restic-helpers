import pytest

import restic_helpers
from restic_helpers.config import Config, TelegramConfig, load_config


@pytest.fixture
def config_dir(tmp_path, monkeypatch):
    monkeypatch.setattr(restic_helpers, "CONFIG_DIR", tmp_path)
    monkeypatch.setattr(restic_helpers, "CONFIG_REPO_DIR", tmp_path / "repositories")
    monkeypatch.setattr("restic_helpers.config.CONFIG_DIR", tmp_path)
    monkeypatch.setattr(
        "restic_helpers.config.CONFIG_REPO_DIR", tmp_path / "repositories"
    )
    return tmp_path


def test_load_empty_config():
    config = Config()
    assert config.telegram.chat_id is None
    assert config.prune.keep_daily == 7


def test_env_override(monkeypatch):
    monkeypatch.setenv("X_RESTIC_TELEGRAM_ENABLED", "false")
    monkeypatch.setenv("X_RESTIC_TELEGRAM_CHAT_ID", "999")
    monkeypatch.setenv("X_RESTIC_TELEGRAM_INFO_BOT_TOKEN", "AAA")
    monkeypatch.setenv("X_RESTIC_TELEGRAM_ERROR_BOT_TOKEN", "BBB")

    config = Config(
        telegram=TelegramConfig(
            enabled=True, chat_id="123", info_bot_token="ZZZ", error_bot_token="YYY"
        )
    )

    assert not config.telegram.enabled
    assert config.telegram.chat_id == "999"
    assert config.telegram.info_bot_token == "AAA"
    assert config.telegram.error_bot_token == "BBB"


def test_secret_override(config_dir):
    (config_dir / "config.toml").write_text(
        '[telegram]\nenabled = true\nchat_id = "123"'
    )
    (config_dir / "secrets.toml").write_text('[telegram]\nchat_id = "456"')

    config = load_config()
    assert config.telegram.enabled
    assert config.telegram.chat_id == "456"


def test_repo_prune_override(config_dir):
    (config_dir / "config.toml").write_text("[prune]\nkeep_daily = 4\nkeep_weekly=2")
    repo_dir = config_dir / "repositories" / "catbook"
    repo_dir.mkdir(parents=True)
    (repo_dir / "prune.toml").write_text("keep_weekly = 8")

    config = load_config()
    assert config.prune.keep_daily == 4
    assert config.repos["catbook"].prune.keep_daily == 4

    assert config.prune.keep_weekly == 2
    assert config.repos["catbook"].prune.keep_weekly == 8

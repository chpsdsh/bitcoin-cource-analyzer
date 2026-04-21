import importlib.util
import json
import sys
from pathlib import Path

import pytest


def load_auth_provider_module(tmp_path: Path, monkeypatch):
    monkeypatch.setenv("AUTH_PROVIDER_DATA_DIR", str(tmp_path))
    module_name = "auth_provider_app_test"
    spec = importlib.util.spec_from_file_location(module_name, Path(__file__).with_name("app.py"))
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def test_create_and_verify_user(tmp_path, monkeypatch):
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)

    auth_app.create_user("User@Example.com", "very-secure-password")

    stored_user = auth_app.get_user("user@example.com")
    assert stored_user is not None
    assert stored_user["email"] == "user@example.com"
    assert auth_app.verify_user("user@example.com", "very-secure-password") is not None
    assert auth_app.verify_user("user@example.com", "wrong-password") is None


def test_create_user_rejects_duplicates(tmp_path, monkeypatch):
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)
    auth_app.create_user("user@example.com", "very-secure-password")

    with pytest.raises(ValueError, match="User already exists"):
        auth_app.create_user("USER@example.com", "very-secure-password")


def test_openid_configuration_uses_public_issuer(tmp_path, monkeypatch):
    monkeypatch.setenv("AUTH_PROVIDER_PUBLIC_ISSUER", "http://localhost:9999")
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)

    config = auth_app.openid_configuration()

    assert config["issuer"] == "http://localhost:9999"
    assert config["authorization_endpoint"].endswith("/authorize")
    assert "RS256" in config["id_token_signing_alg_values_supported"]


def test_issue_tokens_stores_access_token_payload(tmp_path, monkeypatch):
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)
    auth_app.create_user("user@example.com", "very-secure-password")
    user = auth_app.get_user("user@example.com")
    assert user is not None

    tokens = auth_app.issue_tokens(user, "openid profile email")

    assert tokens["token_type"] == "Bearer"
    assert tokens["scope"] == "openid profile email"
    assert tokens["access_token"] in auth_app.access_tokens


def test_hidden_inputs_escapes_quotes(tmp_path, monkeypatch):
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)

    html = auth_app.hidden_inputs({"state": 'quoted"value'})

    assert '&quot;' in html


def test_ensure_users_file_initializes_storage(tmp_path, monkeypatch):
    auth_app = load_auth_provider_module(tmp_path, monkeypatch)

    auth_app.ensure_users_file()

    payload = json.loads((tmp_path / "users.json").read_text())
    assert payload == {"users": []}

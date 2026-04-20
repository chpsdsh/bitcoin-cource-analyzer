import importlib.util
from pathlib import Path

import pytest
from fastapi.testclient import TestClient


@pytest.fixture(scope="session")
def auth_module():
    project_root = Path(__file__).resolve().parents[2]
    app_path = project_root / "auth-provider" / "app.py"

    spec = importlib.util.spec_from_file_location("auth_provider_app", app_path)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


@pytest.fixture(autouse=True)
def isolated_auth_state(auth_module, tmp_path):
    auth_module.sessions.clear()
    auth_module.authorization_codes.clear()
    auth_module.access_tokens.clear()

    auth_module.data_dir = tmp_path
    auth_module.users_path = tmp_path / "users.json"
    auth_module.signing_key_path = tmp_path / "signing-key.pem"

    yield

    auth_module.sessions.clear()
    auth_module.authorization_codes.clear()
    auth_module.access_tokens.clear()


@pytest.fixture
def client(auth_module):
    return TestClient(auth_module.app)


@pytest.fixture
def auth_query(auth_module):
    return {
        "response_type": "code",
        "client_id": auth_module.client_id,
        "redirect_uri": "http://localhost:3000/callback",
        "scope": "openid profile email",
        "state": "test-state",
        "nonce": "test-nonce",
        "code_challenge": "",
        "code_challenge_method": "plain",
    }


@pytest.fixture
def created_user(auth_module):
    email = "user@example.com"
    password = "verysecure123"
    auth_module.create_user(email, password)
    return {"email": email, "password": password}


@pytest.fixture
def logged_in_session(auth_module, created_user):
    session_id = "session-test-123"
    auth_module.sessions[session_id] = created_user["email"]
    return session_id


@pytest.fixture
def issued_code(auth_module, created_user, auth_query):
    code = "auth-code-123"
    auth_module.authorization_codes[code] = {
        "user_email": created_user["email"],
        "redirect_uri": auth_query["redirect_uri"],
        "client_id": auth_query["client_id"],
        "scope": auth_query["scope"],
        "nonce": auth_query["nonce"],
        "code_challenge": auth_query["code_challenge"],
        "code_challenge_method": auth_query["code_challenge_method"],
        "expires_at": 9999999999,
    }
    return code
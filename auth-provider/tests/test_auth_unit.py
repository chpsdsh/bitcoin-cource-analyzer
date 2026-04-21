import base64
import hashlib

import jwt
import pytest
from fastapi import HTTPException
from starlette.requests import Request


def test_password_digest_same_input_same_hash(auth_module):
    salt = "same-salt"
    password = "super-secret-password"

    digest1 = auth_module.password_digest(password, salt)
    digest2 = auth_module.password_digest(password, salt)

    assert digest1 == digest2
    assert digest1 != password
    assert isinstance(digest1, str)


def test_password_digest_different_salt_different_hash(auth_module):
    password = "super-secret-password"

    digest1 = auth_module.password_digest(password, "salt-1")
    digest2 = auth_module.password_digest(password, "salt-2")

    assert digest1 != digest2


def test_create_user_saves_normalized_email_and_hash(auth_module):
    auth_module.create_user("  Test@Example.COM  ", "verysecure123")

    db = auth_module.load_users()
    assert len(db["users"]) == 1

    user = db["users"][0]
    assert user["email"] == "test@example.com"
    assert user["password_hash"] != "verysecure123"
    assert "password_salt" in user
    assert "id" in user
    assert "created_at" in user


def test_create_user_duplicate_email_raises(auth_module):
    auth_module.create_user("test@example.com", "verysecure123")

    with pytest.raises(ValueError, match="User already exists"):
        auth_module.create_user("  TEST@example.com  ", "anotherpass123")


def test_get_user_normalizes_email(auth_module):
    auth_module.create_user("user@example.com", "verysecure123")

    user = auth_module.get_user("  USER@example.com ")

    assert user is not None
    assert user["email"] == "user@example.com"


def test_verify_user_returns_user_for_valid_password(auth_module):
    auth_module.create_user("user@example.com", "verysecure123")

    user = auth_module.verify_user("user@example.com", "verysecure123")

    assert user is not None
    assert user["email"] == "user@example.com"


def test_verify_user_returns_none_for_invalid_password(auth_module):
    auth_module.create_user("user@example.com", "verysecure123")

    user = auth_module.verify_user("user@example.com", "wrong-password")

    assert user is None


def test_verify_user_returns_none_for_unknown_user(auth_module):
    user = auth_module.verify_user("missing@example.com", "whatever123")

    assert user is None


def test_hidden_inputs_contains_all_fields(auth_module, auth_query):
    html = auth_module.hidden_inputs(auth_query)

    for key, value in auth_query.items():
        assert f'name="{key}"' in html
        assert f'value="{value}"' in html


def test_auth_query_from_request_parses_defaults(auth_module):
    scope = {
        "type": "http",
        "method": "GET",
        "path": "/authorize",
        "query_string": b"",
        "headers": [],
    }
    request = Request(scope)

    query = auth_module.auth_query_from_request(request)

    assert query["response_type"] == "code"
    assert query["client_id"] == ""
    assert query["redirect_uri"] == ""
    assert query["scope"] == "openid profile email"
    assert query["state"] == ""
    assert query["nonce"] == ""
    assert query["code_challenge"] == ""
    assert query["code_challenge_method"] == "plain"


def test_validate_client_accepts_valid_credentials(auth_module):
    auth_module.validate_client(auth_module.client_id, auth_module.client_secret)


def test_validate_client_rejects_bad_client_id(auth_module):
    with pytest.raises(HTTPException) as exc:
        auth_module.validate_client("bad-client-id", auth_module.client_secret)

    assert exc.value.status_code == 400
    assert exc.value.detail == "Invalid client_id"


def test_validate_client_rejects_bad_client_secret(auth_module):
    with pytest.raises(HTTPException) as exc:
        auth_module.validate_client(auth_module.client_id, "bad-secret")

    assert exc.value.status_code == 401
    assert exc.value.detail == "Invalid client_secret"


def test_issue_tokens_returns_expected_structure_and_stores_access_token(auth_module):
    auth_module.create_user("user@example.com", "verysecure123")
    user = auth_module.get_user("user@example.com")

    payload = auth_module.issue_tokens(user, "openid profile email")

    assert set(payload.keys()) == {
        "access_token",
        "expires_in",
        "id_token",
        "scope",
        "token_type",
    }
    assert payload["expires_in"] == auth_module.token_lifetime_seconds
    assert payload["scope"] == "openid profile email"
    assert payload["token_type"] == "Bearer"

    access_token = payload["access_token"]
    assert access_token in auth_module.access_tokens
    assert auth_module.access_tokens[access_token]["email"] == "user@example.com"


def test_issue_tokens_id_token_contains_expected_claims(auth_module):
    auth_module.create_user("user@example.com", "verysecure123")
    user = auth_module.get_user("user@example.com")

    payload = auth_module.issue_tokens(user, "openid profile email")

    decoded = jwt.decode(
        payload["id_token"],
        auth_module.public_key,
        algorithms=["RS256"],
        audience=auth_module.client_id,
        issuer=auth_module.public_issuer,
    )

    assert decoded["sub"] == user["id"]
    assert decoded["email"] == "user@example.com"
    assert decoded["email_verified"] is True
    assert decoded["preferred_username"] == "user"
    assert decoded["iss"] == auth_module.public_issuer
    assert decoded["aud"] == auth_module.client_id


def test_register_failed_login_enables_captcha_after_threshold(auth_module):
    key = "127.0.0.1:user@example.com"

    for _ in range(auth_module.captcha_threshold - 1):
        state = auth_module.register_failed_login(key)

    assert state["count"] == auth_module.captcha_threshold - 1
    assert state["captcha_required"] is False

    state = auth_module.register_failed_login(key)

    assert state["count"] == auth_module.captcha_threshold
    assert state["captcha_required"] is True
    assert auth_module.captcha_required_for(key) is True


def test_reset_failed_logins_clears_captcha_requirement(auth_module):
    key = "127.0.0.1:user@example.com"
    for _ in range(auth_module.captcha_threshold):
        auth_module.register_failed_login(key)

    auth_module.reset_failed_logins(key)

    assert auth_module.captcha_required_for(key) is False
    assert key not in auth_module.failed_login_attempts

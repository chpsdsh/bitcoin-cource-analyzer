import base64
import hashlib
from urllib.parse import parse_qs, urlparse

import pytest


def _extract_code_from_location(location: str) -> str:
    parsed = urlparse(location)
    params = parse_qs(parsed.query)
    return params["code"][0]


def test_health_returns_ok(client):
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_openid_configuration_contains_required_fields(client, auth_module):
    response = client.get("/.well-known/openid-configuration")

    assert response.status_code == 200
    payload = response.json()
    assert payload["issuer"] == auth_module.public_issuer
    assert payload["authorization_endpoint"] == f"{auth_module.public_issuer}/authorize"
    assert payload["token_endpoint"] == f"{auth_module.public_issuer}/token"
    assert payload["userinfo_endpoint"] == f"{auth_module.public_issuer}/userinfo"
    assert payload["jwks_uri"] == f"{auth_module.public_issuer}/jwks.json"
    assert "code" in payload["response_types_supported"]
    assert "plain" in payload["code_challenge_methods_supported"]
    assert "S256" in payload["code_challenge_methods_supported"]


def test_jwks_returns_rsa_key(client):
    response = client.get("/jwks.json")

    assert response.status_code == 200
    keys = response.json()["keys"]
    assert len(keys) == 1
    key = keys[0]
    assert key["kty"] == "RSA"
    assert key["kid"] == "btr-local-rs256"
    assert key["alg"] == "RS256"


def test_authorize_requires_valid_client_id(client, auth_query):
    bad_query = {**auth_query, "client_id": "bad-client-id"}

    response = client.get("/authorize", params=bad_query)

    assert response.status_code == 400
    assert response.json()["detail"] == "Invalid client_id"


def test_authorize_requires_redirect_uri(client, auth_query):
    bad_query = {**auth_query, "redirect_uri": ""}

    response = client.get("/authorize", params=bad_query)

    assert response.status_code == 400
    assert response.json()["detail"] == "redirect_uri is required"


def test_authorize_returns_auth_page_for_anonymous_user(client, auth_query):
    response = client.get("/authorize", params=auth_query)

    assert response.status_code == 200
    assert "Create account" in response.text
    assert "Sign in" in response.text
    assert 'action="/login"' in response.text
    assert 'action="/register"' in response.text


def test_authorize_rejects_unsupported_response_type(client, auth_query):
    bad_query = {**auth_query, "response_type": "token"}

    response = client.get("/authorize", params=bad_query)

    assert response.status_code == 400
    assert response.json()["detail"] == "Only authorization code flow is supported"


def test_authorize_redirects_with_code_for_logged_in_user(client, auth_module, auth_query, logged_in_session):
    response = client.get(
        "/authorize",
        params=auth_query,
        cookies={auth_module.session_cookie_name: logged_in_session},
        follow_redirects=False,
    )

    assert response.status_code == 302
    location = response.headers["location"]
    assert location.startswith(auth_query["redirect_uri"])
    assert "code=" in location
    assert f"state={auth_query['state']}" in location

    code = _extract_code_from_location(location)
    assert code in auth_module.authorization_codes


def test_register_success_creates_user_sets_cookie_and_redirects(client, auth_module, auth_query):
    form = {**auth_query, "email": "newuser@example.com", "password": "verysecure123"}

    response = client.post("/register", data=form, follow_redirects=False)

    assert response.status_code == 303
    assert response.headers["location"].startswith("/authorize?")
    assert auth_module.get_user("newuser@example.com") is not None

    set_cookie = response.headers.get("set-cookie", "")
    assert auth_module.session_cookie_name in set_cookie


def test_register_short_password_returns_auth_page_with_error(client, auth_query):
    form = {**auth_query, "email": "newuser@example.com", "password": "short"}

    response = client.post("/register", data=form)

    assert response.status_code == 200
    assert "Password must be at least 8 characters long" in response.text


def test_register_duplicate_user_returns_auth_page_with_error(client, auth_module, auth_query):
    auth_module.create_user("newuser@example.com", "verysecure123")
    form = {**auth_query, "email": "newuser@example.com", "password": "verysecure123"}

    response = client.post("/register", data=form)

    assert response.status_code == 200
    assert "User already exists" in response.text


def test_login_success_sets_cookie_and_redirects(client, auth_module, auth_query, created_user):
    form = {**auth_query, "email": created_user["email"], "password": created_user["password"]}

    response = client.post("/login", data=form, follow_redirects=False)

    assert response.status_code == 303
    assert response.headers["location"].startswith("/authorize?")
    assert auth_module.session_cookie_name in response.headers.get("set-cookie", "")


def test_login_invalid_credentials_returns_error_page(client, auth_query, created_user):
    form = {**auth_query, "email": created_user["email"], "password": "wrong-password"}

    response = client.post("/login", data=form)

    assert response.status_code == 200
    assert "Invalid email or password" in response.text


def test_login_requires_captcha_after_five_failed_attempts(client, auth_module, auth_query, created_user):
    auth_module.captcha_enabled = True
    auth_module.captcha_site_key = "site-key"
    auth_module.captcha_secret_key = "secret-key"
    form = {**auth_query, "email": created_user["email"], "password": "wrong-password"}

    for _ in range(auth_module.captcha_threshold):
        response = client.post("/login", data=form)

    assert response.status_code == 200
    assert "Security check required after repeated failed sign-in attempts." in response.text
    assert "cf-turnstile" in response.text

    missing_captcha_response = client.post("/login", data=form)

    assert missing_captcha_response.status_code == 200
    assert "Please complete the captcha challenge" in missing_captcha_response.text


def test_login_resets_failed_attempts_after_successful_captcha(client, auth_module, auth_query, created_user, monkeypatch):
    auth_module.captcha_enabled = True
    auth_module.captcha_site_key = "site-key"
    auth_module.captcha_secret_key = "secret-key"
    form = {**auth_query, "email": created_user["email"], "password": "wrong-password"}

    for _ in range(auth_module.captcha_threshold):
        client.post("/login", data=form)

    monkeypatch.setattr(auth_module, "verify_captcha", lambda token, remote_ip=None: (True, ""))
    success_form = {
        **auth_query,
        "email": created_user["email"],
        "password": created_user["password"],
        "cf-turnstile-response": "valid-token",
    }

    response = client.post("/login", data=success_form, follow_redirects=False)

    assert response.status_code == 303
    assert response.headers["location"].startswith("/authorize?")

    attempt_key = f"testclient:{created_user['email']}"
    assert attempt_key not in auth_module.failed_login_attempts


def test_token_rejects_unsupported_grant_type(client, auth_module, issued_code, auth_query):
    response = client.post(
        "/token",
        data={
            "grant_type": "refresh_token",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert response.status_code == 400
    assert response.json()["detail"] == "Unsupported grant_type"


def test_token_accepts_client_secret_post(client, auth_module, issued_code, auth_query):
    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["token_type"] == "Bearer"
    assert "access_token" in payload
    assert "id_token" in payload


def test_token_accepts_client_secret_basic(client, auth_module, issued_code, auth_query):
    basic = base64.b64encode(f"{auth_module.client_id}:{auth_module.client_secret}".encode("utf-8")).decode("utf-8")

    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
        },
        headers={"Authorization": f"Basic {basic}"},
    )

    assert response.status_code == 200
    assert response.json()["token_type"] == "Bearer"


def test_token_rejects_invalid_basic_credentials_header(client, auth_query):
    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": "whatever",
            "redirect_uri": auth_query["redirect_uri"],
        },
        headers={"Authorization": "Basic definitely-not-base64"},
    )

    assert response.status_code == 401
    assert response.json()["detail"] == "Invalid client credentials"


def test_token_rejects_missing_client_credentials(client, issued_code, auth_query):
    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
        },
    )

    assert response.status_code == 401
    assert response.json()["detail"] == "Missing client credentials"


def test_token_rejects_invalid_code(client, auth_module, auth_query):
    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": "missing-code",
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert response.status_code == 400
    assert response.json()["detail"] == "Invalid authorization code"


def test_token_rejects_expired_code(client, auth_module, created_user, auth_query):
    code = "expired-code"
    auth_module.authorization_codes[code] = {
        "user_email": created_user["email"],
        "redirect_uri": auth_query["redirect_uri"],
        "client_id": auth_query["client_id"],
        "scope": auth_query["scope"],
        "nonce": auth_query["nonce"],
        "code_challenge": "",
        "code_challenge_method": "plain",
        "expires_at": 0,
    }

    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert response.status_code == 400
    assert response.json()["detail"] == "Authorization code expired"


def test_token_rejects_redirect_uri_mismatch(client, auth_module, issued_code):
    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": "http://localhost:3000/another-callback",
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert response.status_code == 400
    assert response.json()["detail"] == "redirect_uri mismatch"


def test_token_accepts_pkce_plain(client, auth_module, created_user, auth_query):
    code = "pkce-plain-code"
    verifier = "plain-verifier"
    auth_module.authorization_codes[code] = {
        "user_email": created_user["email"],
        "redirect_uri": auth_query["redirect_uri"],
        "client_id": auth_query["client_id"],
        "scope": auth_query["scope"],
        "nonce": auth_query["nonce"],
        "code_challenge": verifier,
        "code_challenge_method": "plain",
        "expires_at": 9999999999,
    }

    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
            "code_verifier": verifier,
        },
    )

    assert response.status_code == 200
    assert "access_token" in response.json()


def test_token_accepts_pkce_s256(client, auth_module, created_user, auth_query):
    code = "pkce-s256-code"
    verifier = "s256-verifier"
    challenge = auth_module.b64url(hashlib.sha256(verifier.encode("utf-8")).digest())

    auth_module.authorization_codes[code] = {
        "user_email": created_user["email"],
        "redirect_uri": auth_query["redirect_uri"],
        "client_id": auth_query["client_id"],
        "scope": auth_query["scope"],
        "nonce": auth_query["nonce"],
        "code_challenge": challenge,
        "code_challenge_method": "S256",
        "expires_at": 9999999999,
    }

    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
            "code_verifier": verifier,
        },
    )

    assert response.status_code == 200
    assert "access_token" in response.json()


def test_token_rejects_wrong_pkce_verifier(client, auth_module, created_user, auth_query):
    code = "bad-pkce-code"
    auth_module.authorization_codes[code] = {
        "user_email": created_user["email"],
        "redirect_uri": auth_query["redirect_uri"],
        "client_id": auth_query["client_id"],
        "scope": auth_query["scope"],
        "nonce": auth_query["nonce"],
        "code_challenge": "expected-verifier",
        "code_challenge_method": "plain",
        "expires_at": 9999999999,
    }

    response = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
            "code_verifier": "wrong-verifier",
        },
    )

    assert response.status_code == 400
    assert response.json()["detail"] == "PKCE verification failed"


def test_token_code_is_one_time_use(client, auth_module, issued_code, auth_query):
    first = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )
    assert first.status_code == 200

    second = client.post(
        "/token",
        data={
            "grant_type": "authorization_code",
            "code": issued_code,
            "redirect_uri": auth_query["redirect_uri"],
            "client_id": auth_module.client_id,
            "client_secret": auth_module.client_secret,
        },
    )

    assert second.status_code == 400
    assert second.json()["detail"] == "Invalid authorization code"


def test_userinfo_requires_bearer_token(client):
    response = client.get("/userinfo")

    assert response.status_code == 401
    assert response.json()["detail"] == "Missing bearer token"


def test_userinfo_rejects_invalid_token(client):
    response = client.get("/userinfo", headers={"Authorization": "Bearer bad-token"})

    assert response.status_code == 401
    assert response.json()["detail"] == "Invalid access token"


def test_userinfo_returns_profile_for_valid_token(client, auth_module, created_user):
    user = auth_module.get_user(created_user["email"])
    issued = auth_module.issue_tokens(user, "openid profile email")
    access_token = issued["access_token"]

    response = client.get("/userinfo", headers={"Authorization": f"Bearer {access_token}"})

    assert response.status_code == 200
    payload = response.json()
    assert payload["sub"] == user["id"]
    assert payload["email"] == created_user["email"]
    assert payload["email_verified"] is True
    assert payload["preferred_username"] == "user"


def test_userinfo_rejects_expired_token(client, auth_module, created_user):
    user = auth_module.get_user(created_user["email"])
    issued = auth_module.issue_tokens(user, "openid profile email")
    access_token = issued["access_token"]

    auth_module.access_tokens[access_token]["exp"] = 0

    response = client.get("/userinfo", headers={"Authorization": f"Bearer {access_token}"})

    assert response.status_code == 401
    assert response.json()["detail"] == "Invalid access token"


def test_logout_clears_session_and_cookie_and_redirects(client, auth_module, logged_in_session):
    response = client.get(
        "/logout",
        params={"rd": "/after-logout"},
        cookies={auth_module.session_cookie_name: logged_in_session},
        follow_redirects=False,
    )

    assert response.status_code == 302
    assert response.headers["location"] == "/after-logout"
    assert logged_in_session not in auth_module.sessions

    set_cookie = response.headers.get("set-cookie", "")
    assert f"{auth_module.session_cookie_name}=" in set_cookie

import base64
import hashlib
import json
import os
import secrets
import time
from pathlib import Path
from typing import Any
from urllib.parse import urlencode

import jwt
from fastapi import FastAPI, Form, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa


app = FastAPI(title="BTR Local OIDC Provider")

data_dir = Path(os.getenv("AUTH_PROVIDER_DATA_DIR", "/data"))
data_dir.mkdir(parents=True, exist_ok=True)
users_path = data_dir / "users.json"
signing_key_path = data_dir / "signing-key.pem"

public_issuer = os.getenv("AUTH_PROVIDER_PUBLIC_ISSUER", "http://localhost:8086")
client_id = os.getenv("AUTH_PROVIDER_CLIENT_ID", "btr-local-client")
client_secret = os.getenv("AUTH_PROVIDER_CLIENT_SECRET", "btr-local-secret")
session_cookie_name = "btr_idp_session"
session_cookie_secure = os.getenv("AUTH_PROVIDER_COOKIE_SECURE", "false").lower() == "true"
token_lifetime_seconds = int(os.getenv("AUTH_PROVIDER_TOKEN_LIFETIME_SECONDS", "3600"))

sessions: dict[str, str] = {}
authorization_codes: dict[str, dict[str, Any]] = {}
access_tokens: dict[str, dict[str, Any]] = {}


def ensure_users_file() -> None:
    if not users_path.exists():
        users_path.write_text(json.dumps({"users": []}, indent=2))


def load_users() -> dict[str, Any]:
    ensure_users_file()
    return json.loads(users_path.read_text())


def save_users(payload: dict[str, Any]) -> None:
    users_path.write_text(json.dumps(payload, indent=2))


def password_digest(password: str, salt: str) -> str:
    return hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt.encode("utf-8"), 200_000).hex()


def create_user(email: str, password: str) -> None:
    db = load_users()
    normalized_email = email.strip().lower()
    if any(user["email"] == normalized_email for user in db["users"]):
        raise ValueError("User already exists")

    salt = secrets.token_hex(16)
    db["users"].append(
        {
            "id": secrets.token_hex(8),
            "email": normalized_email,
            "password_salt": salt,
            "password_hash": password_digest(password, salt),
            "created_at": int(time.time()),
        }
    )
    save_users(db)


def get_user(email: str) -> dict[str, Any] | None:
    db = load_users()
    normalized_email = email.strip().lower()
    return next((user for user in db["users"] if user["email"] == normalized_email), None)


def verify_user(email: str, password: str) -> dict[str, Any] | None:
    user = get_user(email)
    if not user:
        return None

    expected = password_digest(password, user["password_salt"])
    if secrets.compare_digest(expected, user["password_hash"]):
        return user
    return None


def ensure_signing_key() -> rsa.RSAPrivateKey:
    if signing_key_path.exists():
        return serialization.load_pem_private_key(signing_key_path.read_bytes(), password=None)

    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    signing_key_path.write_bytes(
        private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption(),
        )
    )
    return private_key


private_key = ensure_signing_key()
public_key = private_key.public_key()


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode("utf-8").rstrip("=")


def public_jwk() -> dict[str, str]:
    numbers = public_key.public_numbers()
    return {
        "kty": "RSA",
        "use": "sig",
        "alg": "RS256",
        "kid": "btr-local-rs256",
        "n": b64url(numbers.n.to_bytes((numbers.n.bit_length() + 7) // 8, "big")),
        "e": b64url(numbers.e.to_bytes((numbers.e.bit_length() + 7) // 8, "big")),
    }


def validate_client(request_client_id: str, request_client_secret: str | None = None) -> None:
    if request_client_id != client_id:
        raise HTTPException(status_code=400, detail="Invalid client_id")

    if request_client_secret is not None and request_client_secret != client_secret:
        raise HTTPException(status_code=401, detail="Invalid client_secret")


def current_user(request: Request) -> dict[str, Any] | None:
    session_id = request.cookies.get(session_cookie_name)
    if not session_id:
        return None

    email = sessions.get(session_id)
    if not email:
        return None

    return get_user(email)


def hidden_inputs(query: dict[str, Any]) -> str:
    parts = []
    for key, value in query.items():
        safe_value = str(value).replace('"', "&quot;")
        parts.append(f'<input type="hidden" name="{key}" value="{safe_value}">')
    return "\n".join(parts)


def render_auth_page(query: dict[str, Any], error_message: str = "") -> HTMLResponse:
    error_block = f'<div class="error">{error_message}</div>' if error_message else ""
    hidden_fields = hidden_inputs(query)
    html = f"""
    <!doctype html>
    <html lang="en">
      <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>BTR Local Auth</title>
        <style>
          :root {{
            --bg: #0a1118;
            --panel: rgba(16, 23, 34, 0.94);
            --text: #f4f7fb;
            --muted: #92a1b5;
            --accent: #3ad3a6;
            --accent-2: #5a8cff;
            --border: rgba(255,255,255,0.1);
          }}
          * {{ box-sizing: border-box; }}
          body {{
            margin: 0;
            min-height: 100vh;
            display: grid;
            place-items: center;
            padding: 24px;
            font-family: "Space Grotesk", sans-serif;
            color: var(--text);
            background:
              radial-gradient(900px 500px at 5% -10%, rgba(58,211,166,0.16), transparent),
              radial-gradient(900px 700px at 95% 0%, rgba(90,140,255,0.18), transparent),
              linear-gradient(180deg, var(--bg), #05070b);
          }}
          .shell {{
            width: min(980px, 100%);
            display: grid;
            grid-template-columns: 1.1fr 0.9fr;
            gap: 20px;
          }}
          .panel {{
            background: var(--panel);
            border: 1px solid var(--border);
            border-radius: 24px;
            padding: 28px;
            box-shadow: 0 30px 80px rgba(0, 0, 0, 0.35);
          }}
          .hero {{
            display: flex;
            flex-direction: column;
            gap: 18px;
            justify-content: space-between;
          }}
          .eyebrow {{
            color: var(--accent);
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.14em;
          }}
          h1 {{
            margin: 0;
            font-size: clamp(30px, 4vw, 52px);
            line-height: 1.05;
          }}
          p {{
            margin: 0;
            color: var(--muted);
            line-height: 1.6;
          }}
          .form-stack {{
            display: grid;
            gap: 16px;
          }}
          form {{
            display: grid;
            gap: 12px;
          }}
          label {{
            display: grid;
            gap: 6px;
            font-size: 14px;
            color: var(--muted);
          }}
          input {{
            width: 100%;
            padding: 12px 14px;
            border-radius: 14px;
            border: 1px solid var(--border);
            background: rgba(255,255,255,0.03);
            color: var(--text);
          }}
          button {{
            border: 0;
            border-radius: 999px;
            padding: 12px 16px;
            font-weight: 700;
            cursor: pointer;
            background: linear-gradient(135deg, var(--accent), var(--accent-2));
            color: #04111a;
          }}
          .ghost {{
            background: transparent;
            border: 1px solid var(--border);
            color: var(--text);
          }}
          .error {{
            border: 1px solid rgba(255, 86, 86, 0.5);
            background: rgba(255, 86, 86, 0.12);
            color: #ffd0d0;
            border-radius: 14px;
            padding: 12px 14px;
          }}
          .divider {{
            height: 1px;
            background: var(--border);
            margin: 4px 0;
          }}
          @media (max-width: 860px) {{
            .shell {{
              grid-template-columns: 1fr;
            }}
          }}
        </style>
      </head>
      <body>
        <div class="shell">
          <section class="panel hero">
            <div>
              <div class="eyebrow">BTR Access</div>
              <h1>Registration-first OAuth access for the trading dashboard.</h1>
            </div>
            <p>Create an account or sign in to continue. No external database is required here: this local OIDC provider keeps a lightweight user registry for the project demo.</p>
            <p>After successful authentication you will be redirected back to the protected app automatically.</p>
          </section>
          <section class="panel">
            {error_block}
            <div class="form-stack">
              <form method="post" action="/login">
                {hidden_fields}
                <label>Email<input type="email" name="email" required autocomplete="email"></label>
                <label>Password<input type="password" name="password" required autocomplete="current-password"></label>
                <button type="submit">Sign in</button>
              </form>
              <div class="divider"></div>
              <form method="post" action="/register">
                {hidden_fields}
                <label>Email<input type="email" name="email" required autocomplete="email"></label>
                <label>Password<input type="password" name="password" minlength="8" required autocomplete="new-password"></label>
                <button class="ghost" type="submit">Create account</button>
              </form>
            </div>
          </section>
        </div>
      </body>
    </html>
    """
    return HTMLResponse(html)


def auth_query_from_request(request: Request) -> dict[str, Any]:
    return {
        "response_type": request.query_params.get("response_type", "code"),
        "client_id": request.query_params.get("client_id", ""),
        "redirect_uri": request.query_params.get("redirect_uri", ""),
        "scope": request.query_params.get("scope", "openid profile email"),
        "state": request.query_params.get("state", ""),
        "nonce": request.query_params.get("nonce", ""),
        "code_challenge": request.query_params.get("code_challenge", ""),
        "code_challenge_method": request.query_params.get("code_challenge_method", "plain"),
    }


def redirect_to_authorize(query: dict[str, Any]) -> RedirectResponse:
    return RedirectResponse(url=f"/authorize?{urlencode(query)}", status_code=303)


def issue_tokens(user: dict[str, Any], requested_scope: str) -> dict[str, Any]:
    now = int(time.time())
    claims = {
        "iss": public_issuer,
        "sub": user["id"],
        "aud": client_id,
        "email": user["email"],
        "email_verified": True,
        "preferred_username": user["email"].split("@", 1)[0],
        "name": user["email"],
        "iat": now,
        "exp": now + token_lifetime_seconds,
    }
    id_token = jwt.encode(claims, private_key, algorithm="RS256", headers={"kid": "btr-local-rs256"})

    access_token = secrets.token_urlsafe(32)
    access_tokens[access_token] = {
        "sub": user["id"],
        "email": user["email"],
        "scope": requested_scope,
        "exp": now + token_lifetime_seconds,
    }

    return {
        "access_token": access_token,
        "expires_in": token_lifetime_seconds,
        "id_token": id_token,
        "scope": requested_scope,
        "token_type": "Bearer",
    }


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/.well-known/openid-configuration")
def openid_configuration() -> dict[str, Any]:
    return {
        "issuer": public_issuer,
        "authorization_endpoint": f"{public_issuer}/authorize",
        "token_endpoint": f"{public_issuer}/token",
        "userinfo_endpoint": f"{public_issuer}/userinfo",
        "jwks_uri": f"{public_issuer}/jwks.json",
        "response_types_supported": ["code"],
        "subject_types_supported": ["public"],
        "id_token_signing_alg_values_supported": ["RS256"],
        "scopes_supported": ["openid", "profile", "email"],
        "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
        "code_challenge_methods_supported": ["plain", "S256"],
        "claims_supported": ["sub", "iss", "aud", "exp", "iat", "email", "email_verified", "preferred_username", "name"],
    }


@app.get("/jwks.json")
def jwks() -> dict[str, list[dict[str, str]]]:
    return {"keys": [public_jwk()]}


@app.get("/authorize", response_model=None)
def authorize(request: Request) -> HTMLResponse | RedirectResponse:
    query = auth_query_from_request(request)
    validate_client(query["client_id"])

    if query["response_type"] != "code":
        raise HTTPException(status_code=400, detail="Only authorization code flow is supported")

    if not query["redirect_uri"]:
        raise HTTPException(status_code=400, detail="redirect_uri is required")

    user = current_user(request)
    if not user:
        return render_auth_page(query)

    code = secrets.token_urlsafe(32)
    authorization_codes[code] = {
        "user_email": user["email"],
        "redirect_uri": query["redirect_uri"],
        "client_id": query["client_id"],
        "scope": query["scope"],
        "nonce": query["nonce"],
        "code_challenge": query["code_challenge"],
        "code_challenge_method": query["code_challenge_method"] or "plain",
        "expires_at": time.time() + 300,
    }
    redirect_params = {"code": code}
    if query["state"]:
        redirect_params["state"] = query["state"]
    return RedirectResponse(url=f"{query['redirect_uri']}?{urlencode(redirect_params)}", status_code=302)


@app.post("/register", response_model=None)
def register(
    email: str = Form(...),
    password: str = Form(...),
    response_type: str = Form("code"),
    client_id_form: str = Form(..., alias="client_id"),
    redirect_uri: str = Form(...),
    scope: str = Form("openid profile email"),
    state: str = Form(""),
    nonce: str = Form(""),
    code_challenge: str = Form(""),
    code_challenge_method: str = Form("plain"),
) -> HTMLResponse | RedirectResponse:
    query = {
        "response_type": response_type,
        "client_id": client_id_form,
        "redirect_uri": redirect_uri,
        "scope": scope,
        "state": state,
        "nonce": nonce,
        "code_challenge": code_challenge,
        "code_challenge_method": code_challenge_method,
    }

    try:
        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters long")
        create_user(email, password)
    except ValueError as error:
        return render_auth_page(query, str(error))

    session_id = secrets.token_urlsafe(24)
    sessions[session_id] = email.strip().lower()
    response = redirect_to_authorize(query)
    response.set_cookie(
        session_cookie_name,
        session_id,
        httponly=True,
        samesite="lax",
        secure=session_cookie_secure,
        path="/",
    )
    return response


@app.post("/login", response_model=None)
def login(
    email: str = Form(...),
    password: str = Form(...),
    response_type: str = Form("code"),
    client_id_form: str = Form(..., alias="client_id"),
    redirect_uri: str = Form(...),
    scope: str = Form("openid profile email"),
    state: str = Form(""),
    nonce: str = Form(""),
    code_challenge: str = Form(""),
    code_challenge_method: str = Form("plain"),
) -> HTMLResponse | RedirectResponse:
    query = {
        "response_type": response_type,
        "client_id": client_id_form,
        "redirect_uri": redirect_uri,
        "scope": scope,
        "state": state,
        "nonce": nonce,
        "code_challenge": code_challenge,
        "code_challenge_method": code_challenge_method,
    }

    user = verify_user(email, password)
    if not user:
        return render_auth_page(query, "Invalid email or password")

    session_id = secrets.token_urlsafe(24)
    sessions[session_id] = user["email"]
    response = redirect_to_authorize(query)
    response.set_cookie(
        session_cookie_name,
        session_id,
        httponly=True,
        samesite="lax",
        secure=session_cookie_secure,
        path="/",
    )
    return response


@app.post("/token")
def token(
    request: Request,
    grant_type: str = Form(...),
    code: str = Form(...),
    redirect_uri: str = Form(...),
    client_id_form: str | None = Form(None, alias="client_id"),
    client_secret_form: str | None = Form(None, alias="client_secret"),
    code_verifier: str = Form(""),
) -> JSONResponse:
    if grant_type != "authorization_code":
        raise HTTPException(status_code=400, detail="Unsupported grant_type")

    authorization_header = request.headers.get("Authorization", "")
    if authorization_header.startswith("Basic "):
        try:
            decoded = base64.b64decode(authorization_header.split(" ", 1)[1]).decode("utf-8")
            client_id_form, client_secret_form = decoded.split(":", 1)
        except Exception as error:
            raise HTTPException(status_code=401, detail="Invalid client credentials") from error

    if not client_id_form or client_secret_form is None:
        raise HTTPException(status_code=401, detail="Missing client credentials")

    validate_client(client_id_form, client_secret_form)

    code_payload = authorization_codes.pop(code, None)
    if not code_payload:
        raise HTTPException(status_code=400, detail="Invalid authorization code")

    if time.time() > code_payload["expires_at"]:
        raise HTTPException(status_code=400, detail="Authorization code expired")

    if redirect_uri != code_payload["redirect_uri"]:
        raise HTTPException(status_code=400, detail="redirect_uri mismatch")

    method = code_payload["code_challenge_method"]
    stored_challenge = code_payload["code_challenge"]
    if stored_challenge:
        if method == "S256":
            calculated = b64url(hashlib.sha256(code_verifier.encode("utf-8")).digest())
        else:
            calculated = code_verifier
        if calculated != stored_challenge:
            raise HTTPException(status_code=400, detail="PKCE verification failed")

    user = get_user(code_payload["user_email"])
    if not user:
        raise HTTPException(status_code=400, detail="Unknown user")

    return JSONResponse(issue_tokens(user, code_payload["scope"]))


@app.get("/userinfo")
def userinfo(request: Request) -> JSONResponse:
    authorization = request.headers.get("Authorization", "")
    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing bearer token")

    token = authorization.split(" ", 1)[1]
    token_payload = access_tokens.get(token)
    if not token_payload or token_payload["exp"] < time.time():
        raise HTTPException(status_code=401, detail="Invalid access token")

    return JSONResponse(
        {
            "sub": token_payload["sub"],
            "email": token_payload["email"],
            "email_verified": True,
            "preferred_username": token_payload["email"].split("@", 1)[0],
            "name": token_payload["email"],
        }
    )


@app.get("/logout")
def logout(request: Request, rd: str = "/") -> RedirectResponse:
    session_id = request.cookies.get(session_cookie_name)
    if session_id:
        sessions.pop(session_id, None)

    response = RedirectResponse(url=rd, status_code=302)
    response.delete_cookie(session_cookie_name, path="/")
    return response

# bitcoin-cource-analyzer

## Automated build and run

Use one script for common project operations:

```bash
./scripts/project.sh up
```

Available commands:

```bash
./scripts/project.sh up
./scripts/project.sh down
./scripts/project.sh restart
./scripts/project.sh status
./scripts/project.sh logs
./scripts/project.sh logs web-frontend
```

After startup:

- Frontend: `http://localhost:8085`
- Local OIDC provider: `http://localhost:8086`
- Kafka UI: `http://localhost:8091`

## OAuth Access

The frontend is protected by `oauth2-proxy` in front of Nginx.
Unauthenticated users are redirected to the OAuth provider before they can open the site, load news, or call prediction endpoints.

Default behavior:

- `./scripts/project.sh up` works without any extra setup.
- The stack includes a built-in local OIDC provider with registration and login pages.
- Open `http://localhost:8085`, create an account, and the app will continue after OAuth login.
- User accounts are stored in a lightweight local JSON-backed volume, not in an external database.

Optional external provider setup:

- Copy `.env.example` to `.env`
- Replace the default local OIDC values with your external provider values
- Keep the callback URL as:

```text
http://localhost:8085/oauth2/callback
```

Generate a cookie secret for external provider usage with:

```bash
python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(32)).decode())'
```

Security notes:

- `news-gateway`, `llm-consumer`, and `llm` are no longer published to host ports, so app functionality is available only through the authenticated frontend.
- The built-in local OIDC provider is intended for project/dev usage. For production, override it with a real external OIDC provider in `.env`.

## Frontend E2E

The repository now includes Cypress-based E2E tests for the static frontend.

Install dependencies:

```bash
npm install
```

Run the suite:

```bash
npm run test:e2e
```

Force Chrome on macOS:

```bash
npm run test:e2e:chrome
```

Force WebKit on macOS:

```bash
npm run test:e2e:webkit
```

Open Cypress locally:

```bash
npm run cy:open
```

The Cypress config starts a lightweight local static server for `web-frontend/html`, so Docker services are not required for the frontend E2E suite. API calls are stubbed with Cypress intercepts.

On macOS, the Cypress launcher prefers installed browsers like Chrome or Firefox. If they are not available, it falls back to Cypress WebKit support, which makes the setup friendlier for Apple Silicon MacBooks that only have Safari/WebKit available by default.


## CI

GitHub Actions contains one pipeline:

- `CI` in `.github/workflows/ci.yml`

### CI

On `pull_request` and push to `main`, the pipeline:

- runs `golangci-lint`, `go test ./...`, and `go build` for each Go service
- runs Python tests for `llm` and `auth-provider`, then checks syntax with `python -m compileall`
- validates the frontend Dockerfile build
- validates `docker-compose.yml`
- runs SonarQube analysis for the monorepo and imports Go and Python coverage reports

## SonarQube

The repository includes a dedicated `SonarQube Scan` job in GitHub Actions.

Before it can work, configure these repository settings:

- Repository secret `SONAR_HOST_URL` with your SonarQube server URL, for example `https://sonarqube.example.com`
- Repository secret `SONAR_TOKEN` with a token that can run analysis for this project
- Repository variable `SONAR_PROJECT_KEY` with the SonarQube project key
- Optional repository variable `SONAR_PROJECT_NAME` if you want a display name different from the repository name

The scan uses [`sonar-project.properties`](/Users/andrewf1amex/Programming/bitcoin-cource-analyzer/sonar-project.properties:1) and currently imports coverage for:

- `cache-service`
- `llm-consumer`
- `news-gateway`
- `news-parser`
- `llm`
- `auth-provider`

The frontend is still analyzed without coverage import for now.

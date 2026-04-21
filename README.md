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
- installs Python dependencies for `llm` and checks syntax with `python -m compileall`
- validates the frontend Dockerfile build
- validates `docker-compose.yml`
- runs JMeter-based performance checks for `news-gateway` and `llm-consumer`

## JMeter Performance Tests

The repository now includes JMeter plans and CI jobs for backend load checks:

- `news-smoke`: category read smoke test for `GET /news/:category`
- `news-hot-category`: burst load against a hot category such as `crypto`
- `news-mixed-profile`: weighted category mix closer to real read traffic
- `news-cache-cold-hot`: cold-cache pass followed by a stricter hot-cache pass
- `llm-consumer-predict`: concurrent `POST /predict` load against `llm-consumer`

Helper files:

- JMeter plans live in `jmeter/`
- the CI-only Binance stub lives in `perf/mock-binance/`
- `docker-compose.perf.yml` overrides `llm-consumer` to use the local mock price feed

Run a plan locally with Dockerized JMeter:

```bash
chmod +x scripts/run-jmeter.sh scripts/assert-jmeter-results.sh
COMPOSE_PROJECT_NAME=btca-local docker compose -f docker-compose.yml -f docker-compose.perf.yml up -d --build valkey mock-binance news-gateway llm-consumer
COMPOSE_PROJECT_NAME=btca-local docker compose -f docker-compose.yml -f docker-compose.perf.yml run --rm valkey-seed
COMPOSE_PROJECT_NAME=btca-local ./scripts/run-jmeter.sh jmeter/news-gateway-read-csv.jmx artifacts/news-smoke-local -Jhost=news-gateway -Jport=8080 -Jthreads=5 -Jramp_up=1 -Jloops=4 -Jcategories_file=jmeter/data/news-categories-smoke.csv
./scripts/assert-jmeter-results.sh artifacts/news-smoke-local/results.jtl 20 0 2000 1000
COMPOSE_PROJECT_NAME=btca-local docker compose -f docker-compose.yml -f docker-compose.perf.yml down -v
```

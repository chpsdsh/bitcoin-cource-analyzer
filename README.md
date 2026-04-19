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
- News gateway: `http://localhost:8080`
- LLM consumer: `http://localhost:8083`
- LLM service: `http://localhost:8084`
- Kafka UI: `http://localhost:8091`

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

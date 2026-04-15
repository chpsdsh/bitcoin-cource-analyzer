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


## CI

GitHub Actions contains one pipeline:

- `CI` in `.github/workflows/ci.yml`

### CI

On `pull_request` and push to `main`, the pipeline:

- runs `golangci-lint`, `go test ./...`, and `go build` for each Go service
- installs Python dependencies for `llm` and checks syntax with `python -m compileall`
- validates the frontend Dockerfile build
- validates `docker-compose.yml`

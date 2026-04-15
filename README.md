# bitcoin-cource-analyzer

## CI

GitHub Actions contains one pipeline:

- `CI` in `.github/workflows/ci.yml`

### CI

On `pull_request` and push to `main`, the pipeline:

- runs `golangci-lint`, `go test ./...`, and `go build` for each Go service
- installs Python dependencies for `llm` and checks syntax with `python -m compileall`
- validates the frontend Dockerfile build
- validates `docker-compose.yml`

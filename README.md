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

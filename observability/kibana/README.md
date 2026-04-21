Kibana is available at http://localhost:5601 after `docker compose up`.

Create a data view for `app-logs-*` and use `@timestamp` as the time field.
Filter by `trace_id` to follow a request or background parsing cycle across services.

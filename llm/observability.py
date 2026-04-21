import json
import logging
import sys
from contextvars import ContextVar
from datetime import datetime, timezone


TRACE_ID_HEADER = "X-Request-ID"
current_trace_id: ContextVar[str] = ContextVar("current_trace_id", default="")


class JsonFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        payload = {
            "time": datetime.now(timezone.utc).isoformat(),
            "level": record.levelname,
            "service": "llm",
            "msg": record.getMessage(),
        }

        for key, value in record.__dict__.items():
            if key.startswith("_"):
                payload[key[1:]] = value

        if "trace_id" not in payload:
            trace_id = current_trace_id.get()
            if trace_id:
                payload["trace_id"] = trace_id

        if record.exc_info:
            payload["error"] = self.formatException(record.exc_info)

        return json.dumps(payload, ensure_ascii=False)


def configure_logging() -> None:
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JsonFormatter())

    root = logging.getLogger()
    root.handlers.clear()
    root.addHandler(handler)
    root.setLevel(logging.INFO)


logger = logging.getLogger("llm")


def set_trace_id(trace_id: str) -> None:
    current_trace_id.set(trace_id)

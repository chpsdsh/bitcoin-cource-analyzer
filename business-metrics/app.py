from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Lock
from urllib.parse import urlparse


PORT = 8080


class MetricsState:
    def __init__(self):
        self.lock = Lock()
        self.day = self._today()
        self.visits = 0
        self.predictions = 0
        self.users = set()
        self.visit_hours = [0] * 24

    def _today(self):
        return datetime.now(timezone.utc).date().isoformat()

    def _reset_if_needed(self):
        today = self._today()
        if today == self.day:
            return
        self.day = today
        self.visits = 0
        self.predictions = 0
        self.users.clear()
        self.visit_hours = [0] * 24

    def observe(self, email, uri):
        parsed = urlparse(uri or "/")
        path = parsed.path or "/"
        if not self._is_page_visit(path) and path != "/predict":
            return

        now = datetime.now(timezone.utc)
        user = (email or "anonymous").strip().lower()
        with self.lock:
            self._reset_if_needed()
            if self._is_page_visit(path):
                self.visits += 1
                self.users.add(user)
                self.visit_hours[now.hour] += 1
            if path == "/predict":
                self.predictions += 1
                self.users.add(user)

    def render_prometheus(self):
        with self.lock:
            self._reset_if_needed()
            unique_users = len(self.users)
            visits_per_user = self.visits / unique_users if unique_users else 0
            predictions_per_user = self.predictions / unique_users if unique_users else 0
            lines = [
                "# HELP business_visits_today_total Authenticated visits since 00:00 UTC.",
                "# TYPE business_visits_today_total gauge",
                f'business_visits_today_total{{day="{self.day}"}} {self.visits}',
                "# HELP business_unique_visitors_today Authenticated unique visitors since 00:00 UTC.",
                "# TYPE business_unique_visitors_today gauge",
                f'business_unique_visitors_today{{day="{self.day}"}} {unique_users}',
                "# HELP business_predictions_today_total Prediction requests since 00:00 UTC.",
                "# TYPE business_predictions_today_total gauge",
                f'business_predictions_today_total{{day="{self.day}"}} {self.predictions}',
                "# HELP business_avg_predictions_per_user_today Average prediction requests per unique visitor since 00:00 UTC.",
                "# TYPE business_avg_predictions_per_user_today gauge",
                f'business_avg_predictions_per_user_today{{day="{self.day}"}} {predictions_per_user:.6f}',
                "# HELP business_avg_visits_per_user_today Average visits per unique visitor since 00:00 UTC.",
                "# TYPE business_avg_visits_per_user_today gauge",
                f'business_avg_visits_per_user_today{{day="{self.day}"}} {visits_per_user:.6f}',
                "# HELP business_visit_density_today Visits grouped by UTC hour for the current day.",
                "# TYPE business_visit_density_today gauge",
            ]
            for hour, count in enumerate(self.visit_hours):
                lines.append(f'business_visit_density_today{{day="{self.day}",hour="{hour:02d}:00"}} {count}')
            return "\n".join(lines) + "\n"

    def _is_page_visit(self, path):
        return path in ("/", "/index.html")


state = MetricsState()


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/metrics":
            body = state.render_prometheus().encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return
        if self.path == "/visit":
            state.observe(self.headers.get("X-Email"), self.headers.get("X-Original-Uri"))
        self.send_response(204)
        self.end_headers()

    def do_POST(self):
        if self.path == "/visit":
            state.observe(self.headers.get("X-Email"), self.headers.get("X-Original-Uri"))
        self.send_response(204)
        self.end_headers()

    def log_message(self, format, *args):
        return


if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", PORT), Handler).serve_forever()

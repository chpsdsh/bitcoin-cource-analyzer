from fastapi import BackgroundTasks

import app as llm_app
from schemas import Features, ScoreRequest, SummarizeRequest


def test_health_reports_current_model_state(monkeypatch):
    monkeypatch.setattr(llm_app.llm_service, "is_loaded", True)

    response = llm_app.health()

    assert response.status == "ok"
    assert response.model_loaded is True
    assert response.model_name == llm_app.settings.model_name


def test_summarize_uses_service_and_settings(monkeypatch):
    captured = {}

    def fake_summarize(category, news, max_chars_per_news, max_new_tokens):
        captured.update(
            {
                "category": category,
                "news": news,
                "max_chars_per_news": max_chars_per_news,
                "max_new_tokens": max_new_tokens,
            }
        )
        return {
            "category": category,
            "summarization": "summary",
            "features": {
                "signal_direction": "up",
                "signal_strength": 0.8,
                "uncertainty": 0.2,
                "event_urgency_hours": 6,
                "numbers_density": 1.2,
                "entity_density": 0.7,
            },
        }

    monkeypatch.setattr(llm_app.llm_service, "summarize", fake_summarize)
    request = SummarizeRequest(category="btc", news=["one", "two"], max_new_tokens=256)

    response = llm_app.summarize(request)

    assert response.category == "btc"
    assert response.summarization == "summary"
    assert response.features.signal_direction == "up"
    assert captured == {
        "category": "btc",
        "news": ["one", "two"],
        "max_chars_per_news": llm_app.settings.max_input_chars_per_news,
        "max_new_tokens": 256,
    }


def test_score_uses_serialized_features(monkeypatch):
    captured = {}

    def fake_score(category, summarization, features, max_new_tokens):
        captured.update(
            {
                "category": category,
                "summarization": summarization,
                "features": features,
                "max_new_tokens": max_new_tokens,
            }
        )
        return {
            "category": category,
            "score": 0.55,
            "verdict": "growth",
            "rationale_ru": "positive",
        }

    monkeypatch.setattr(llm_app.llm_service, "score", fake_score)
    request = ScoreRequest(
        category="btc",
        summarization="summary",
        features=Features(
            signal_direction="up",
            signal_strength=0.5,
            uncertainty=0.1,
            event_urgency_hours=4,
            numbers_density=0.3,
            entity_density=0.2,
        ),
    )

    response = llm_app.score(request)

    assert response.score == 0.55
    assert response.verdict == "growth"
    assert captured["category"] == "btc"
    assert captured["summarization"] == "summary"
    assert captured["features"]["signal_direction"] == "up"
    assert captured["max_new_tokens"] == llm_app.settings.score_max_new_tokens


def test_full_pipeline_runs_both_stages(monkeypatch):
    def fake_summarize(**kwargs):
        return {
            "category": kwargs["category"],
            "summarization": "summary",
            "features": {"signal_direction": "neutral"},
        }

    def fake_score(**kwargs):
        return {
            "category": kwargs["category"],
            "score": 0.0,
            "verdict": "neutral",
            "rationale_ru": "balanced",
        }

    monkeypatch.setattr(llm_app.llm_service, "summarize", fake_summarize)
    monkeypatch.setattr(llm_app.llm_service, "score", fake_score)

    response = llm_app.full_pipeline(SummarizeRequest(category="macro", news=["headline"]))

    assert response["summary_stage"]["summarization"] == "summary"
    assert response["score_stage"]["verdict"] == "neutral"
    assert response["model_version"] == llm_app.settings.model_version


def test_predict_rejects_parallel_run(monkeypatch):
    background_tasks = BackgroundTasks()
    monkeypatch.setattr(llm_app, "_prediction_job_running", True)

    response = llm_app.predict(background_tasks)

    assert response == {"status": "already_running"}


def test_predict_accepts_and_schedules_job(monkeypatch):
    background_tasks = BackgroundTasks()
    monkeypatch.setattr(llm_app, "_prediction_job_running", False)

    response = llm_app.predict(background_tasks)

    assert response == {"status": "accepted"}
    assert llm_app._prediction_job_running is True
    assert len(background_tasks.tasks) == 1

from unittest.mock import PropertyMock, patch


def test_health_returns_service_state(client, llm_app_module):
    with patch.object(type(llm_app_module.llm_service), "is_loaded", new_callable=PropertyMock) as mocked:
        mocked.return_value = True

        response = client.get("/health")

    assert response.status_code == 200
    payload = response.json()
    assert payload["status"] == "ok"
    assert payload["model_loaded"] is True
    assert payload["model_name"] == llm_app_module.settings.model_name


def test_version_returns_versions(client, llm_app_module):
    response = client.get("/version")

    assert response.status_code == 200
    payload = response.json()
    assert payload["service_version"] == "0.1.0"
    assert payload["model_version"] == llm_app_module.settings.model_version
    assert payload["prompt_version"] == llm_app_module.settings.prompt_version


def test_summarize_endpoint_returns_service_result(
    client,
    llm_app_module,
    monkeypatch,
    summary_result_payload,
):
    calls = {}

    def fake_summarize(category, news, max_chars_per_news, max_new_tokens):
        calls["category"] = category
        calls["news"] = news
        calls["max_chars_per_news"] = max_chars_per_news
        calls["max_new_tokens"] = max_new_tokens
        return summary_result_payload

    monkeypatch.setattr(llm_app_module.llm_service, "summarize", fake_summarize)

    response = client.post(
        "/summarize",
        json={
            "category": "macro",
            "news": [" First news ", "Second news"],
            "max_new_tokens": 256,
        },
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["category"] == summary_result_payload["category"]
    assert payload["summarization"] == summary_result_payload["summarization"]
    assert payload["features"] == summary_result_payload["features"]
    assert payload["model_version"] == llm_app_module.settings.model_version
    assert payload["prompt_version"] == llm_app_module.settings.prompt_version

    assert calls["category"] == "macro"
    assert calls["news"] == ["First news", "Second news"]
    assert calls["max_chars_per_news"] == llm_app_module.settings.max_input_chars_per_news
    assert calls["max_new_tokens"] == 256


def test_summarize_endpoint_returns_500_on_service_error(client, llm_app_module, monkeypatch):
    def fake_summarize(*args, **kwargs):
        raise RuntimeError("summary exploded")

    monkeypatch.setattr(llm_app_module.llm_service, "summarize", fake_summarize)

    response = client.post(
        "/summarize",
        json={
            "category": "macro",
            "news": ["news 1"],
        },
    )

    assert response.status_code == 500
    assert "summarize failed: summary exploded" in response.json()["detail"]


def test_score_endpoint_returns_service_result(
    client,
    llm_app_module,
    monkeypatch,
    score_result_payload,
):
    calls = {}

    def fake_score(category, summarization, features, max_new_tokens):
        calls["category"] = category
        calls["summarization"] = summarization
        calls["features"] = features
        calls["max_new_tokens"] = max_new_tokens
        return score_result_payload

    monkeypatch.setattr(llm_app_module.llm_service, "score", fake_score)

    response = client.post(
        "/score",
        json={
            "category": "macro",
            "summarization": "RU summary",
            "features": {
                "signal_direction": "up",
                "signal_strength": 0.7,
                "uncertainty": 0.2,
                "event_urgency_hours": 6,
                "numbers_density": 0.3,
                "entity_density": 0.4,
            },
            "max_new_tokens": 300,
        },
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["category"] == score_result_payload["category"]
    assert payload["score"] == score_result_payload["score"]
    assert payload["verdict"] == score_result_payload["verdict"]
    assert payload["rationale_ru"] == score_result_payload["rationale_ru"]
    assert payload["model_version"] == llm_app_module.settings.model_version
    assert payload["prompt_version"] == llm_app_module.settings.prompt_version

    assert calls["category"] == "macro"
    assert calls["summarization"] == "RU summary"
    assert calls["features"]["signal_direction"] == "up"
    assert calls["max_new_tokens"] == 300


def test_score_endpoint_returns_500_on_service_error(client, llm_app_module, monkeypatch):
    def fake_score(*args, **kwargs):
        raise RuntimeError("score exploded")

    monkeypatch.setattr(llm_app_module.llm_service, "score", fake_score)

    response = client.post(
        "/score",
        json={
            "category": "macro",
            "summarization": "RU summary",
            "features": {
                "signal_direction": "up",
            },
        },
    )

    assert response.status_code == 500
    assert "score failed: score exploded" in response.json()["detail"]


def test_pipeline_calls_summarize_then_score_and_merges_response(
    client,
    llm_app_module,
    monkeypatch,
    summary_result_payload,
    score_result_payload,
):
    summarize_calls = []
    score_calls = []

    def fake_summarize(category, news, max_chars_per_news, max_new_tokens):
        summarize_calls.append(
            {
                "category": category,
                "news": news,
                "max_chars_per_news": max_chars_per_news,
                "max_new_tokens": max_new_tokens,
            }
        )
        return summary_result_payload

    def fake_score(category, summarization, features, max_new_tokens):
        score_calls.append(
            {
                "category": category,
                "summarization": summarization,
                "features": features,
                "max_new_tokens": max_new_tokens,
            }
        )
        return score_result_payload

    monkeypatch.setattr(llm_app_module.llm_service, "summarize", fake_summarize)
    monkeypatch.setattr(llm_app_module.llm_service, "score", fake_score)

    response = client.post(
        "/pipeline",
        json={
            "category": "macro",
            "news": ["news 1", "news 2"],
        },
    )

    assert response.status_code == 200
    payload = response.json()

    assert payload["summary_stage"] == summary_result_payload
    assert payload["score_stage"] == score_result_payload
    assert payload["model_version"] == llm_app_module.settings.model_version
    assert payload["prompt_version"] == llm_app_module.settings.prompt_version

    assert len(summarize_calls) == 1
    assert len(score_calls) == 1
    assert score_calls[0]["category"] == summary_result_payload["category"]
    assert score_calls[0]["summarization"] == summary_result_payload["summarization"]
    assert score_calls[0]["features"] == summary_result_payload["features"]
    assert score_calls[0]["max_new_tokens"] == llm_app_module.settings.score_max_new_tokens


def test_pipeline_returns_500_when_service_fails(client, llm_app_module, monkeypatch):
    def fake_summarize(*args, **kwargs):
        raise RuntimeError("pipeline exploded")

    monkeypatch.setattr(llm_app_module.llm_service, "summarize", fake_summarize)

    response = client.post(
        "/pipeline",
        json={
            "category": "macro",
            "news": ["news 1"],
        },
    )

    assert response.status_code == 500
    assert "pipeline failed: pipeline exploded" in response.json()["detail"]


def test_predict_starts_background_job_when_not_running(client, llm_app_module, monkeypatch):
    llm_app_module._prediction_job_running = False

    called = {"value": False}

    def fake_run_prediction_job_guarded():
        called["value"] = True

    monkeypatch.setattr(llm_app_module, "_run_prediction_job_guarded", fake_run_prediction_job_guarded)

    response = client.post("/predict")

    assert response.status_code == 200
    assert response.json() == {"status": "accepted"}
    assert called["value"] is True


def test_predict_returns_already_running_when_job_is_locked(client, llm_app_module):
    llm_app_module._prediction_job_running = True

    response = client.post("/predict")

    assert response.status_code == 200
    assert response.json() == {"status": "already_running"}

    llm_app_module._prediction_job_running = False


def test_run_prediction_job_guarded_resets_running_flag_on_exception(llm_app_module, monkeypatch):
    llm_app_module._prediction_job_running = True

    def fake_run_prediction_job():
        raise RuntimeError("background failed")

    monkeypatch.setattr(llm_app_module, "run_prediction_job", fake_run_prediction_job)

    try:
        llm_app_module._run_prediction_job_guarded()
    except RuntimeError:
        pass

    assert llm_app_module._prediction_job_running is False
import json
from types import SimpleNamespace

def test_extract_news_text_prefers_fields_in_expected_order(jobs_module):
    assert jobs_module._extract_news_text({"text": " text value "}) == "text value"
    assert jobs_module._extract_news_text({"content": " content value "}) == "content value"
    assert jobs_module._extract_news_text({"summary": " summary value "}) == "summary value"
    assert jobs_module._extract_news_text({"title": " title value "}) == "title value"


def test_extract_news_text_returns_empty_string_when_no_fields(jobs_module):
    assert jobs_module._extract_news_text({}) == ""
    assert jobs_module._extract_news_text({"text": "   "}) == ""


def test_read_latest_news_by_category_collects_valid_items(jobs_module, monkeypatch):
    class FakeRedis:
        def scan_iter(self, pattern):
            assert pattern == "*"
            return iter(["macro", "mining"])

        def zrevrange(self, key, start, end):
            assert start == 0
            assert end == 1
            if key == "macro":
                return [
                    json.dumps({"text": "Macro news 1"}),
                    json.dumps({"content": "Macro news 2"}),
                ]
            if key == "mining":
                return [
                    json.dumps({"summary": "Mining summary"}),
                ]
            return []

    monkeypatch.setattr(jobs_module, "build_redis_client", lambda: FakeRedis())

    result = jobs_module._read_latest_news_by_category(limit=2)

    assert result == {
        "macro": ["Macro news 1", "Macro news 2"],
        "mining": ["Mining summary"],
    }


def test_read_latest_news_by_category_skips_invalid_json_and_empty_text(jobs_module, monkeypatch):
    class FakeRedis:
        def scan_iter(self, pattern):
            return iter(["macro"])

        def zrevrange(self, key, start, end):
            return [
                "not-json-at-all",
                json.dumps({"text": "   "}),
                json.dumps({"title": "Valid title"}),
            ]

    monkeypatch.setattr(jobs_module, "build_redis_client", lambda: FakeRedis())

    result = jobs_module._read_latest_news_by_category(limit=3)

    assert result == {"macro": ["Valid title"]}


def test_run_prediction_job_processes_categories_and_sends_events(
    jobs_module,
    monkeypatch,
    summary_result_payload,
    score_result_payload,
):
    sent_messages = []
    flush_called = {"value": False}
    summarize_calls = []
    score_calls = []

    class FakeProducer:
        def send(self, topic, key, value):
            sent_messages.append({"topic": topic, "key": key, "value": value})

        def flush(self):
            flush_called["value"] = True

    def fake_summarize(category, news, max_chars_per_news, max_new_tokens):
        summarize_calls.append(
            {
                "category": category,
                "news": news,
                "max_chars_per_news": max_chars_per_news,
                "max_new_tokens": max_new_tokens,
            }
        )
        return {
            **summary_result_payload,
            "category": category,
        }

    def fake_score(category, summarization, features, max_new_tokens):
        score_calls.append(
            {
                "category": category,
                "summarization": summarization,
                "features": features,
                "max_new_tokens": max_new_tokens,
            }
        )
        return {
            **score_result_payload,
            "category": category,
        }

    fake_settings = SimpleNamespace(
        kafka_llm_response_topic="llm_data",
        model_version=jobs_module.settings.model_version,
        prompt_version=jobs_module.settings.prompt_version,
        llm_top_n=2,
        max_input_chars_per_news=jobs_module.settings.max_input_chars_per_news,
        default_max_new_tokens=jobs_module.settings.default_max_new_tokens,
        score_max_new_tokens=jobs_module.settings.score_max_new_tokens,
    )

    monkeypatch.setattr(
        jobs_module,
        "_read_latest_news_by_category",
        lambda limit: {
            "macro": ["news 1", "news 2"],
            "etf": ["news 3"],
        },
    )
    monkeypatch.setattr(jobs_module, "build_kafka_producer", lambda: FakeProducer())
    monkeypatch.setattr(jobs_module.llm_service, "summarize", fake_summarize)
    monkeypatch.setattr(jobs_module.llm_service, "score", fake_score)
    monkeypatch.setattr(jobs_module, "settings", fake_settings)

    jobs_module.run_prediction_job()

    assert len(summarize_calls) == 2
    assert len(score_calls) == 2
    assert len(sent_messages) == 2
    assert flush_called["value"] is True

    first_message = sent_messages[0]
    assert first_message["topic"] == "llm_data"
    assert first_message["key"] in {"macro", "etf"}

    event = first_message["value"]
    assert "timestamp_utc" in event
    assert event["category"] in {"macro", "etf"}
    assert event["summarization"] == summary_result_payload["summarization"]
    assert event["features"] == summary_result_payload["features"]
    assert event["score"] == score_result_payload["score"]
    assert event["verdict"] == score_result_payload["verdict"]
    assert event["rationale_ru"] == score_result_payload["rationale_ru"]
    assert event["model_version"] == fake_settings.model_version
    assert event["prompt_version"] == fake_settings.prompt_version


def test_run_prediction_job_continues_when_one_category_fails(
    jobs_module,
    monkeypatch,
    summary_result_payload,
    score_result_payload,
):
    sent_messages = []
    flush_called = {"value": False}

    class FakeProducer:
        def send(self, topic, key, value):
            sent_messages.append({"topic": topic, "key": key, "value": value})

        def flush(self):
            flush_called["value"] = True

    def fake_summarize(category, news, max_chars_per_news, max_new_tokens):
        if category == "broken":
            raise RuntimeError("boom")
        return {
            **summary_result_payload,
            "category": category,
        }

    def fake_score(category, summarization, features, max_new_tokens):
        return {
            **score_result_payload,
            "category": category,
        }

    monkeypatch.setattr(
        jobs_module,
        "_read_latest_news_by_category",
        lambda limit: {
            "broken": ["bad news"],
            "macro": ["good news"],
        },
    )
    monkeypatch.setattr(jobs_module, "build_kafka_producer", lambda: FakeProducer())
    monkeypatch.setattr(jobs_module.llm_service, "summarize", fake_summarize)
    monkeypatch.setattr(jobs_module.llm_service, "score", fake_score)

    jobs_module.run_prediction_job()

    assert len(sent_messages) == 1
    assert sent_messages[0]["key"] == "macro"
    assert flush_called["value"] is True
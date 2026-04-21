import pytest
from pydantic import ValidationError


def test_summarize_request_accepts_valid_payload():
    from schemas import SummarizeRequest

    payload = SummarizeRequest(
        category="macro",
        news=[" first news ", "second news"],
        max_new_tokens=256,
    )

    assert payload.category == "macro"
    assert payload.news == ["first news", "second news"]
    assert payload.max_new_tokens == 256


def test_summarize_request_rejects_empty_category():
    from schemas import SummarizeRequest

    with pytest.raises(ValidationError):
        SummarizeRequest(
            category="",
            news=["news 1"],
        )


def test_summarize_request_rejects_empty_news_list():
    from schemas import SummarizeRequest

    with pytest.raises(ValidationError):
        SummarizeRequest(
            category="macro",
            news=[],
        )


def test_summarize_request_rejects_all_empty_news_items():
    from schemas import SummarizeRequest

    with pytest.raises(ValidationError, match="news must contain at least one non-empty item"):
        SummarizeRequest(
            category="macro",
            news=["   ", "\n\n"],
        )


def test_summarize_request_limits_max_new_tokens_range():
    from schemas import SummarizeRequest

    with pytest.raises(ValidationError):
        SummarizeRequest(
            category="macro",
            news=["news 1"],
            max_new_tokens=64,
        )


def test_features_accepts_valid_payload():
    from schemas import Features

    features = Features(
        signal_direction="up",
        signal_strength=0.8,
        uncertainty=0.2,
        event_urgency_hours=12,
        numbers_density=0.4,
        entity_density=0.3,
    )

    assert features.signal_direction == "up"
    assert features.signal_strength == 0.8
    assert features.uncertainty == 0.2
    assert features.event_urgency_hours == 12
    assert features.numbers_density == 0.4
    assert features.entity_density == 0.3


def test_features_rejects_signal_strength_out_of_range():
    from schemas import Features

    with pytest.raises(ValidationError):
        Features(
            signal_direction="up",
            signal_strength=1.5,
        )


def test_features_rejects_uncertainty_out_of_range():
    from schemas import Features

    with pytest.raises(ValidationError):
        Features(
            signal_direction="neutral",
            uncertainty=-0.1,
        )


def test_features_rejects_negative_density():
    from schemas import Features

    with pytest.raises(ValidationError):
        Features(
            signal_direction="down",
            numbers_density=-0.01,
        )


def test_score_request_accepts_valid_payload():
    from schemas import ScoreRequest

    req = ScoreRequest(
        category="macro",
        summarization="Краткая сводка",
        features={
            "signal_direction": "up",
            "signal_strength": 0.7,
        },
        max_new_tokens=256,
    )

    assert req.category == "macro"
    assert req.summarization == "Краткая сводка"
    assert req.features.signal_direction == "up"
    assert req.max_new_tokens == 256


def test_score_request_rejects_empty_summarization():
    from schemas import ScoreRequest

    with pytest.raises(ValidationError):
        ScoreRequest(
            category="macro",
            summarization="",
            features={"signal_direction": "up"},
        )


def test_score_result_accepts_valid_payload():
    from schemas import ScoreResult

    result = ScoreResult(
        category="macro",
        score=0.42,
        verdict="growth",
        rationale_ru="Позитивный сигнал",
        model_version="qwen",
        prompt_version="v1",
    )

    assert result.score == 0.42
    assert result.verdict == "growth"
    assert result.mode == "score"


def test_score_result_rejects_score_out_of_range():
    from schemas import ScoreResult

    with pytest.raises(ValidationError):
        ScoreResult(
            category="macro",
            score=1.2,
            verdict="growth",
            rationale_ru="text",
            model_version="qwen",
            prompt_version="v1",
        )


def test_score_result_rejects_invalid_verdict():
    from schemas import ScoreResult

    with pytest.raises(ValidationError):
        ScoreResult(
            category="macro",
            score=0.0,
            verdict="sideways",
            rationale_ru="text",
            model_version="qwen",
            prompt_version="v1",
        )
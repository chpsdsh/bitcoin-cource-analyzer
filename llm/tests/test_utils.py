import pytest


def test_trim_text_collapses_whitespace():
    from utils import trim_text

    text = "  hello \n\n   world\t test  "
    result = trim_text(text, max_chars=100)

    assert result == "hello world test"


def test_trim_text_respects_max_chars():
    from utils import trim_text

    text = "abcdefghij"
    result = trim_text(text, max_chars=5)

    assert result == "abcde"


def test_build_news_block_formats_enumerated_items():
    from utils import build_news_block

    news = [" First news  ", "Second\n\nnews"]
    result = build_news_block(news, max_chars_per_news=100)

    assert result == "1) First news\n2) Second news"


def test_safe_json_loads_parses_plain_json():
    from utils import safe_json_loads

    result = safe_json_loads('{"a": 1, "b": "x"}')

    assert result == {"a": 1, "b": "x"}


def test_safe_json_loads_parses_fenced_json():
    from utils import safe_json_loads

    result = safe_json_loads(
        """```json
        {"a": 1, "b": "x"}
        ```"""
    )

    assert result == {"a": 1, "b": "x"}


def test_safe_json_loads_raises_on_invalid_json():
    from utils import safe_json_loads

    with pytest.raises(Exception):
        safe_json_loads("not a json")


def test_verdict_from_score_returns_growth():
    from utils import verdict_from_score

    assert verdict_from_score(0.16) == "growth"


def test_verdict_from_score_returns_drop():
    from utils import verdict_from_score

    assert verdict_from_score(-0.16) == "drop"


def test_verdict_from_score_returns_neutral_inside_threshold():
    from utils import verdict_from_score

    assert verdict_from_score(0.15) == "neutral"
    assert verdict_from_score(-0.15) == "neutral"
    assert verdict_from_score(0.0) == "neutral"
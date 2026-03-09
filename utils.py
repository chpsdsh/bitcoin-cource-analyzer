import json
import re
from typing import Any


def trim_text(text: str, max_chars: int) -> str:
    text = re.sub(r"\s+", " ", text).strip()
    return text[:max_chars]


def build_news_block(news: list[str], max_chars_per_news: int) -> str:
    blocks = []
    for i, item in enumerate(news, start=1):
        blocks.append(f"{i}) {trim_text(item, max_chars_per_news)}")
    return "\n".join(blocks)


def safe_json_loads(text: str) -> dict[str, Any]:
    text = text.strip()

    # На случай если модель обернула в ```json ... ```
    text = re.sub(r"^```json\s*", "", text)
    text = re.sub(r"^```\s*", "", text)
    text = re.sub(r"\s*```$", "", text)

    return json.loads(text)


def verdict_from_score(score: float) -> str:
    if score > 0.15:
        return "growth"
    if score < -0.15:
        return "drop"
    return "neutral"
import json
from datetime import datetime, timezone

from config import settings
from infra import build_kafka_producer, build_redis_client
from llm import llm_service


def _extract_news_text(item: dict) -> str:
    return (
        item.get("text")
        or item.get("content")
        or item.get("summary")
        or item.get("title")
        or ""
    ).strip()


def _read_latest_news_by_category(limit: int) -> dict[str, list[str]]:
    redis_client = build_redis_client()
    result: dict[str, list[str]] = {}

    for key in redis_client.scan_iter("*"):
        raw_items = redis_client.zrevrange(key, 0, limit - 1)
        news_texts: list[str] = []

        for raw in raw_items:
            try:
                payload = json.loads(raw)
                text = _extract_news_text(payload)
                if text:
                    news_texts.append(text)
            except Exception:
                continue

        if news_texts:
            result[key] = news_texts

    return result


def run_prediction_job() -> None:
    producer = build_kafka_producer()
    category_to_news = _read_latest_news_by_category(settings.llm_top_n)

    for category, news_list in category_to_news.items():
        try:
            summary_result = llm_service.summarize(
                category=category,
                news=news_list,
                max_chars_per_news=settings.max_input_chars_per_news,
                max_new_tokens=settings.default_max_new_tokens,
            )

            score_result = llm_service.score(
                category=summary_result["category"],
                summarization=summary_result["summarization"],
                features=summary_result["features"],
                max_new_tokens=300,
            )

            event = {
                "timestamp_utc": datetime.now(timezone.utc).isoformat(),
                "category": summary_result["category"],
                "summarization": summary_result["summarization"],
                "features": summary_result["features"],
                "score": score_result["score"],
                "verdict": score_result["verdict"],
                "rationale_ru": score_result["rationale_ru"],
                "news_count": len(news_list),
                "model_version": settings.model_version,
                "prompt_version": settings.prompt_version,
            }

            producer.send(
                settings.kafka_llm_response_topic,
                key=category,
                value=event,
            )

        except Exception as e:
            print(f"[predict-job] failed for category={category}: {e}", flush=True)

    producer.flush()
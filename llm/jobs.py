import json
import logging
from datetime import datetime, timezone
import time

from config import settings
from infra import build_kafka_producer, build_redis_client
from llm import llm_service
from observability import set_trace_id

logger = logging.getLogger("llm")


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
    scanned_keys = 0

    for key in redis_client.scan_iter("*"):
        scanned_keys += 1
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

    if scanned_keys == 0:
        logger.info("redis returned no keys")
    elif not result:
        logger.info("redis keys found, but no news items were extracted")

    return result


def run_prediction_job(trace_id: str = "") -> None:
    set_trace_id(trace_id)
    producer = build_kafka_producer()
    category_to_news = _read_latest_news_by_category(settings.llm_top_n)
    logger.info("prediction job loaded categories", extra={"_trace_id": trace_id, "_categories_count": len(category_to_news)})

    for category, news_list in category_to_news.items():
        try:
            category_started_at = time.perf_counter()
            logger.info(
                "prediction category processing started",
                extra={"_trace_id": trace_id, "_category": category, "_news_count": len(news_list)},
            )
            summarize_started_at = time.perf_counter()
            logger.info("summarize started", extra={"_trace_id": trace_id, "_category": category})
            summary_result = llm_service.summarize(
                category=category,
                news=news_list,
                max_chars_per_news=settings.max_input_chars_per_news,
                max_new_tokens=settings.default_max_new_tokens,
            )
            logger.info(
                "summarize completed",
                extra={
                    "_trace_id": trace_id,
                    "_category": category,
                    "_elapsed_sec": round(time.perf_counter() - summarize_started_at, 2),
                },
            )

            score_started_at = time.perf_counter()
            logger.info("score started", extra={"_trace_id": trace_id, "_category": category})
            score_result = llm_service.score(
                category=summary_result["category"],
                summarization=summary_result["summarization"],
                features=summary_result["features"],
                max_new_tokens=settings.score_max_new_tokens,
            )
            logger.info(
                "score completed",
                extra={
                    "_trace_id": trace_id,
                    "_category": category,
                    "_elapsed_sec": round(time.perf_counter() - score_started_at, 2),
                },
            )

            event = {
                "trace_id": trace_id,
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
                headers=[("trace_id", trace_id.encode("utf-8"))] if trace_id else None,
            )
            logger.info(
                "prediction sent",
                extra={
                    "_trace_id": trace_id,
                    "_category": category,
                    "_news_count": len(news_list),
                    "_total_elapsed_sec": round(time.perf_counter() - category_started_at, 2),
                },
            )

        except Exception as e:
            logger.exception("prediction failed", extra={"_trace_id": trace_id, "_category": category, "_error": str(e)})

    producer.flush()
    logger.info("producer flush completed", extra={"_trace_id": trace_id})

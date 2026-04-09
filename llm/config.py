from dataclasses import dataclass
import os


@dataclass(frozen=True)
class Settings:
    model_name: str = os.getenv("MODEL_NAME", "Qwen/Qwen2.5-0.5B-Instruct")
    model_version: str = os.getenv("MODEL_VERSION", "qwen2.5-0.5b-instruct")
    prompt_version: str = os.getenv("PROMPT_VERSION", "v1")

    default_max_new_tokens: int = int(os.getenv("DEFAULT_MAX_NEW_TOKENS", "128"))
    repair_max_new_tokens: int = int(os.getenv("REPAIR_MAX_NEW_TOKENS", "96"))
    score_max_new_tokens: int = int(os.getenv("SCORE_MAX_NEW_TOKENS", "128"))
    max_input_chars_per_news: int = int(os.getenv("MAX_INPUT_CHARS_PER_NEWS", "1000"))

    min_news_count: int = int(os.getenv("MIN_NEWS_COUNT", "1"))
    max_news_count: int = int(os.getenv("MAX_NEWS_COUNT", "7"))

    redis_addr: str = os.getenv("REDIS_ADDR", "valkey:6379")
    redis_password: str = os.getenv("REDIS_PASSWORD", "")
    redis_llm_response_db: int = int(os.getenv("REDIS_LLM_RESPONSE_DB", "0"))

    kafka_brokers: str = os.getenv("KAFKA_BROKERS", "kafka:19092")
    kafka_llm_response_topic: str = os.getenv("KAFKA_LLM_RESPONSE_TOPIC", "llm_data")

    llm_top_n: int = int(os.getenv("LLM_TOP_N", "2"))


settings = Settings()

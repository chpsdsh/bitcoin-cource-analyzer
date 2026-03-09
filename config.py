from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    model_name: str = "Qwen/Qwen2.5-7B-Instruct"
    model_version: str = "qwen2.5-7b-instruct"
    prompt_version: str = "v1"
    default_max_new_tokens: int = 700
    repair_max_new_tokens: int = 500
    max_input_chars_per_news: int = 5000
    min_news_count: int = 1
    max_news_count: int = 7


settings = Settings()
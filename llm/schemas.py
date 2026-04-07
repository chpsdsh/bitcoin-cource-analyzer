from typing import Any, Literal

from pydantic import BaseModel, Field, field_validator


SignalDirection = Literal["up", "down", "neutral"]


class SummarizeRequest(BaseModel):
    category: str = Field(..., min_length=1, max_length=100)
    news: list[str] = Field(..., min_length=1, max_length=7)
    max_new_tokens: int | None = Field(default=None, ge=128, le=2048)

    @field_validator("news")
    @classmethod
    def validate_news(cls, news: list[str]) -> list[str]:
        cleaned = [item.strip() for item in news if item and item.strip()]
        if not cleaned:
            raise ValueError("news must contain at least one non-empty item")
        return cleaned


class Features(BaseModel):
    signal_direction: SignalDirection
    signal_strength: float | None = Field(default=None, ge=0.0, le=1.0)
    uncertainty: float | None = Field(default=None, ge=0.0, le=1.0)
    event_urgency_hours: int | None = Field(default=None)
    numbers_density: float | None = Field(default=None, ge=0.0)
    entity_density: float | None = Field(default=None, ge=0.0)


class SummarizeResponse(BaseModel):
    category: str
    summarization: str
    features: Features
    model_version: str
    prompt_version: str
    mode: Literal["summarize"] = "summarize"


class ScoreRequest(BaseModel):
    category: str
    summarization: str = Field(..., min_length=1)
    features: Features
    max_new_tokens: int | None = Field(default=None, ge=128, le=1024)


class ScoreResult(BaseModel):
    category: str
    score: float = Field(..., ge=-1.0, le=1.0)
    verdict: Literal["growth", "neutral", "drop"]
    rationale_ru: str
    model_version: str
    prompt_version: str
    mode: Literal["score"] = "score"


class HealthResponse(BaseModel):
    status: str
    model_loaded: bool
    model_name: str


class ErrorResponse(BaseModel):
    detail: str


class RawLLMResponse(BaseModel):
    raw_text: str
    parsed_json: dict[str, Any]
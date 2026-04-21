import traceback
import threading
from fastapi import FastAPI, HTTPException, BackgroundTasks, Request
from jobs import run_prediction_job
from observability import TRACE_ID_HEADER, configure_logging, logger

from config import settings
from llm import llm_service
from schemas import (
    ErrorResponse,
    HealthResponse,
    ScoreRequest,
    ScoreResult,
    SummarizeRequest,
    SummarizeResponse,
)

configure_logging()

app = FastAPI(
    title="BTC ML Service Skeleton",
    version="0.1.0",
    description="LLM microservice for category summarization and scoring",
)


_prediction_job_lock = threading.Lock()
_prediction_job_running = False


def _run_prediction_job_guarded(trace_id: str = "") -> None:
    global _prediction_job_running
    try:
        if trace_id:
            run_prediction_job(trace_id)
        else:
            run_prediction_job()
    finally:
        with _prediction_job_lock:
            _prediction_job_running = False


@app.get("/health", response_model=HealthResponse)
def health() -> HealthResponse:
    return HealthResponse(
        status="ok",
        model_loaded=llm_service.is_loaded,
        model_name=settings.model_name,
    )


@app.get("/version")
def version() -> dict[str, str]:
    return {
        "service_version": "0.1.0",
        "model_version": settings.model_version,
        "prompt_version": settings.prompt_version,
    }


@app.post(
    "/summarize",
    response_model=SummarizeResponse,
    responses={400: {"model": ErrorResponse}, 500: {"model": ErrorResponse}},
)
def summarize(request: SummarizeRequest) -> SummarizeResponse:
    try:
        result = llm_service.summarize(
            category=request.category,
            news=request.news,
            max_chars_per_news=settings.max_input_chars_per_news,
            max_new_tokens=request.max_new_tokens or settings.default_max_new_tokens,
        )
        return SummarizeResponse(
            category=result["category"],
            summarization=result["summarization"],
            features=result["features"],
            model_version=settings.model_version,
            prompt_version=settings.prompt_version,
        )
    except Exception as exc:
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=f"summarize failed: {exc}") from exc



@app.post(
    "/score",
    response_model=ScoreResult,
    responses={400: {"model": ErrorResponse}, 500: {"model": ErrorResponse}},
)
def score(request: ScoreRequest) -> ScoreResult:
    try:
        result = llm_service.score(
            category=request.category,
            summarization=request.summarization,
            features=request.features.model_dump(),
            max_new_tokens=request.max_new_tokens or settings.score_max_new_tokens,
        )
        return ScoreResult(
            category=result["category"],
            score=result["score"],
            verdict=result["verdict"],
            rationale_ru=result["rationale_ru"],
            model_version=settings.model_version,
            prompt_version=settings.prompt_version,
        )
    except Exception as exc:
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=f"score failed: {exc}") from exc


@app.post("/pipeline")
def full_pipeline(request: SummarizeRequest) -> dict:
    try:
        summary_result = llm_service.summarize(
            category=request.category,
            news=request.news,
            max_chars_per_news=settings.max_input_chars_per_news,
            max_new_tokens=request.max_new_tokens or settings.default_max_new_tokens,
        )

        score_result = llm_service.score(
            category=summary_result["category"],
            summarization=summary_result["summarization"],
            features=summary_result["features"],
            max_new_tokens=settings.score_max_new_tokens,
        )

        return {
            "summary_stage": summary_result,
            "score_stage": score_result,
            "model_version": settings.model_version,
            "prompt_version": settings.prompt_version,
        }
    except Exception as exc:
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=f"pipeline failed: {exc}") from exc
    
    
def predict(background_tasks: BackgroundTasks, trace_id: str = "") -> dict[str, str]:
    global _prediction_job_running

    with _prediction_job_lock:
        if _prediction_job_running:
            logger.info("prediction job already running", extra={"_trace_id": trace_id})
            return {"status": "already_running"}
        _prediction_job_running = True

    logger.info("prediction job accepted", extra={"_trace_id": trace_id})

    def run_guarded() -> None:
        if trace_id:
            _run_prediction_job_guarded(trace_id)
            return
        _run_prediction_job_guarded()

    background_tasks.add_task(run_guarded)
    return {"status": "accepted"}


@app.post("/predict")
def predict_endpoint(request: Request, background_tasks: BackgroundTasks) -> dict[str, str]:
    return predict(background_tasks, request.headers.get(TRACE_ID_HEADER, ""))

import sys
from pathlib import Path

import pytest
from fastapi.testclient import TestClient


@pytest.fixture(scope="session", autouse=True)
def add_llm_to_syspath():
    project_root = Path(__file__).resolve().parents[2]
    llm_dir = project_root / "llm"

    llm_dir_str = str(llm_dir)
    if llm_dir_str not in sys.path:
        sys.path.insert(0, llm_dir_str)

    yield

    if llm_dir_str in sys.path:
        sys.path.remove(llm_dir_str)


@pytest.fixture
def llm_app_module(add_llm_to_syspath):
    import app

    return app


@pytest.fixture
def jobs_module(add_llm_to_syspath):
    import jobs

    return jobs


@pytest.fixture
def infra_module(add_llm_to_syspath):
    import infra

    return infra


@pytest.fixture
def client(llm_app_module):
    return TestClient(llm_app_module.app)


@pytest.fixture
def summary_result_payload():
    return {
        "category": "macro",
        "summarization": "Краткая сводка по макроэкономическим новостям.",
        "features": {
            "signal_direction": "up",
            "signal_strength": 0.8,
            "uncertainty": 0.2,
            "event_urgency_hours": 12,
            "numbers_density": 0.4,
            "entity_density": 0.3,
        },
    }


@pytest.fixture
def score_result_payload():
    return {
        "category": "macro",
        "score": 0.42,
        "verdict": "growth",
        "rationale_ru": "Позитивный макросигнал для BTC.",
    }
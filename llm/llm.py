from __future__ import annotations

import threading
import time
from typing import Any

import torch
from transformers import AutoModelForCausalLM, AutoTokenizer

from config import settings
from observability import logger
from prompts import SCORING_SYSTEM_PROMPT, SUMMARIZE_SYSTEM_PROMPT
from utils import build_news_block, safe_json_loads


class LLMService:
    def __init__(self) -> None:
        self.model = None
        self.tokenizer = None
        self.model_name = settings.model_name
        self.device = "cpu"
        self._load_lock = threading.Lock()

    def _resolve_device(self) -> tuple[str, torch.dtype]:
        if torch.cuda.is_available():
            return "cuda", torch.float16
        if torch.backends.mps.is_available():
            return "mps", torch.float16
        return "cpu", torch.float32

    @property
    def is_loaded(self) -> bool:
        return self.model is not None and self.tokenizer is not None

    def load(self) -> None:
        if self.is_loaded:
            return

        with self._load_lock:
            if self.is_loaded:
                return

            self.device, torch_dtype = self._resolve_device()
            logger.info(
                "model loading started",
                extra={"_model": self.model_name, "_device": self.device, "_dtype": str(torch_dtype)},
            )
            started_at = time.perf_counter()
            logger.info("tokenizer load started", extra={"_model": self.model_name})
            self.tokenizer = AutoTokenizer.from_pretrained(self.model_name, use_fast=True)
            logger.info(
                "tokenizer load completed",
                extra={"_model": self.model_name, "_elapsed_sec": round(time.perf_counter() - started_at, 2)},
            )
            model_load_started_at = time.perf_counter()
            logger.info("model weights load started", extra={"_model": self.model_name})
            self.model = AutoModelForCausalLM.from_pretrained(
                self.model_name,
                torch_dtype=torch_dtype,
                low_cpu_mem_usage=True,
            ).to(self.device)
            self.model.eval()
            logger.info(
                "model weights load completed",
                extra={"_model": self.model_name, "_elapsed_sec": round(time.perf_counter() - model_load_started_at, 2)},
            )
            logger.info(
                "model loaded",
                extra={"_model": self.model_name, "_total_elapsed_sec": round(time.perf_counter() - started_at, 2)},
            )

    def _generate_text(
        self,
        system_prompt: str,
        user_prompt: str,
        max_new_tokens: int,
    ) -> str:
        self.load()
        generation_started_at = time.perf_counter()
        logger.info(
            "generation started",
            extra={"_max_new_tokens": max_new_tokens, "_device": self.device},
        )
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ]
        
        text = self.tokenizer.apply_chat_template(
            messages,
            tokenize=False,
            add_generation_prompt=True,
        )
        model_inputs = self.tokenizer(text, return_tensors="pt")
        model_inputs = model_inputs.to(self.model.device)
        
        with torch.inference_mode():
            generated = self.model.generate(
                **model_inputs,
                max_new_tokens=max_new_tokens,
                do_sample=False,
                repetition_penalty=1.03,
                eos_token_id=self.tokenizer.eos_token_id,
                pad_token_id=self.tokenizer.eos_token_id,
            )

        generated_ids = generated[0, model_inputs["input_ids"].shape[1]:]
        response = self.tokenizer.decode(generated_ids, skip_special_tokens=True).strip()
        logger.info(
            "generation completed",
            extra={
                "_elapsed_sec": round(time.perf_counter() - generation_started_at, 2),
                "_output_chars": len(response),
            },
        )
        return response

    def _repair_json(
        self,
        broken_response: str,
        original_user_prompt: str,
        system_prompt: str,
        max_new_tokens: int,
    ) -> dict[str, Any]:
        repair_started_at = time.perf_counter()
        logger.info("json repair started")
        self.load()

        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": original_user_prompt},
            {"role": "assistant", "content": broken_response},
            {
                "role": "user",
                "content": (
                    "Исправь предыдущий ответ так, чтобы он был строго валидным JSON по схеме. "
                    "Не добавляй новых фактов. Верни только JSON."
                ),
            },
        ]

        text = self.tokenizer.apply_chat_template(
            messages,
            tokenize=False,
            add_generation_prompt=True,
        )
        model_inputs = self.tokenizer(text, return_tensors="pt").to(self.model.device)

        with torch.inference_mode():
            generated = self.model.generate(
                **model_inputs,
                max_new_tokens=max_new_tokens,
                do_sample=False,
                repetition_penalty=1.03,
                eos_token_id=self.tokenizer.eos_token_id,
                pad_token_id=self.tokenizer.eos_token_id,
            )

        generated_ids = generated[0, model_inputs["input_ids"].shape[1]:]
        fixed = self.tokenizer.decode(generated_ids, skip_special_tokens=True).strip()
        logger.info(
            "json repair completed",
            extra={
                "_elapsed_sec": round(time.perf_counter() - repair_started_at, 2),
                "_output_chars": len(fixed),
            },
        )
        return safe_json_loads(fixed)

    def summarize(
        self,
        category: str,
        news: list[str],
        max_chars_per_news: int,
        max_new_tokens: int,
    ) -> dict[str, Any]:
        user_prompt = (
            f"Category: {category}\n\n"
            f"News (English):\n{build_news_block(news, max_chars_per_news)}"
        )
        raw = self._generate_text(
            system_prompt=SUMMARIZE_SYSTEM_PROMPT,
            user_prompt=user_prompt,
            max_new_tokens=max_new_tokens,
        )
        try:
            return safe_json_loads(raw)
        except Exception:
            return self._repair_json(
                broken_response=raw,
                original_user_prompt=user_prompt,
                system_prompt=SUMMARIZE_SYSTEM_PROMPT,
                max_new_tokens=settings.repair_max_new_tokens,
            )

    def score(
        self,
        category: str,
        summarization: str,
        features: dict[str, Any],
        max_new_tokens: int,
    ) -> dict[str, Any]:
        user_prompt = (
            f"Category: {category}\n\n"
            f"Summarization (RU):\n{summarization}\n\n"
            f"Features:\n{features}"
        )

        raw = self._generate_text(
            system_prompt=SCORING_SYSTEM_PROMPT,
            user_prompt=user_prompt,
            max_new_tokens=max_new_tokens,
        )

        try:
            return safe_json_loads(raw)
        except Exception:
            return self._repair_json(
                broken_response=raw,
                original_user_prompt=user_prompt,
                system_prompt=SCORING_SYSTEM_PROMPT,
                max_new_tokens=settings.repair_max_new_tokens,
            )


llm_service = LLMService()

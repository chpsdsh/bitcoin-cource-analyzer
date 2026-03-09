from __future__ import annotations

from typing import Any

import torch
from transformers import AutoModelForCausalLM, AutoTokenizer

from config import settings
from prompts import SCORING_SYSTEM_PROMPT, SUMMARIZE_SYSTEM_PROMPT
from utils import build_news_block, safe_json_loads


class LLMService:
    def __init__(self) -> None:
        self.model = None
        self.tokenizer = None
        self.model_name = settings.model_name

    @property
    def is_loaded(self) -> bool:
        return self.model is not None and self.tokenizer is not None

    def load(self) -> None:
        if self.is_loaded:
            return

        self.tokenizer = AutoTokenizer.from_pretrained(self.model_name, use_fast=True)
        self.model = AutoModelForCausalLM.from_pretrained(
            self.model_name,
            torch_dtype=torch.float16,
        ).to("cuda")
        self.model.eval()

    def _generate_text(
        self,
        system_prompt: str,
        user_prompt: str,
        max_new_tokens: int,
    ) -> str:
        print("[1] before load()", flush=True)
        self.load()
        print("[2] after load()", flush=True)

        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ]

        print("[3] before apply_chat_template", flush=True)
        text = self.tokenizer.apply_chat_template(
            messages,
            tokenize=False,
            add_generation_prompt=True,
        )
        print("[4] after apply_chat_template", flush=True)

        print("[5] before tokenizer()", flush=True)
        model_inputs = self.tokenizer(text, return_tensors="pt")
        print("[6] after tokenizer()", flush=True)

        print("[7] before to(device)", flush=True)
        model_inputs = model_inputs.to(self.model.device)
        print("[8] after to(device)", flush=True)

        print("[9] before generate()", flush=True)
        with torch.inference_mode():
            generated = self.model.generate(
                **model_inputs,
                max_new_tokens=max_new_tokens,
                do_sample=False,
                temperature=0.0,
                top_p=1.0,
                repetition_penalty=1.03,
                eos_token_id=self.tokenizer.eos_token_id,
                pad_token_id=self.tokenizer.eos_token_id,
            )
        print("[10] after generate()", flush=True)

        generated_ids = generated[0, model_inputs["input_ids"].shape[1]:]
        response = self.tokenizer.decode(generated_ids, skip_special_tokens=True).strip()
        print("[11] decoded response", flush=True)
        return response

    def _repair_json(
        self,
        broken_response: str,
        original_user_prompt: str,
        system_prompt: str,
        max_new_tokens: int,
    ) -> dict[str, Any]:
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
                temperature=0.0,
                top_p=1.0,
                repetition_penalty=1.03,
                eos_token_id=self.tokenizer.eos_token_id,
                pad_token_id=self.tokenizer.eos_token_id,
            )

        generated_ids = generated[0, model_inputs["input_ids"].shape[1]:]
        fixed = self.tokenizer.decode(generated_ids, skip_special_tokens=True).strip()
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
            print("TRIED")
            return safe_json_loads(raw)
        except Exception:
            print("EXCEPTED")
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
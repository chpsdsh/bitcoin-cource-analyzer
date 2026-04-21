import torch


class FakeBatch(dict):
    def to(self, device):
        return self


class FakeTokenizer:
    def __init__(self):
        self.eos_token_id = 42
        self.chat_calls = []
        self.tokenize_calls = []
        self.decode_calls = []

    def apply_chat_template(self, messages, tokenize, add_generation_prompt):
        self.chat_calls.append(
            {
                "messages": messages,
                "tokenize": tokenize,
                "add_generation_prompt": add_generation_prompt,
            }
        )
        return "PROMPT_TEXT"

    def __call__(self, text, return_tensors):
        self.tokenize_calls.append(
            {
                "text": text,
                "return_tensors": return_tensors,
            }
        )
        return FakeBatch(
            {
                "input_ids": torch.tensor([[10, 11, 12]]),
                "attention_mask": torch.tensor([[1, 1, 1]]),
            }
        )

    def decode(self, generated_ids, skip_special_tokens):
        self.decode_calls.append(
            {
                "generated_ids": generated_ids,
                "skip_special_tokens": skip_special_tokens,
            }
        )
        return ' {"ok": true} '


class FakeModel:
    def __init__(self):
        self.device = "cpu"
        self.to_calls = []
        self.eval_called = False
        self.generate_calls = []

    def to(self, device):
        self.device = device
        self.to_calls.append(device)
        return self

    def eval(self):
        self.eval_called = True

    def generate(self, **kwargs):
        self.generate_calls.append(kwargs)
        return torch.tensor([[10, 11, 12, 21, 22]])


def test_resolve_device_prefers_cuda(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    monkeypatch.setattr(torch.cuda, "is_available", lambda: True)
    monkeypatch.setattr(torch.backends.mps, "is_available", lambda: False)

    service = LLMService()
    device, dtype = service._resolve_device()

    assert device == "cuda"
    assert dtype == torch.float16


def test_resolve_device_uses_mps_when_cuda_unavailable(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    monkeypatch.setattr(torch.cuda, "is_available", lambda: False)
    monkeypatch.setattr(torch.backends.mps, "is_available", lambda: True)

    service = LLMService()
    device, dtype = service._resolve_device()

    assert device == "mps"
    assert dtype == torch.float16


def test_resolve_device_falls_back_to_cpu(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    monkeypatch.setattr(torch.cuda, "is_available", lambda: False)
    monkeypatch.setattr(torch.backends.mps, "is_available", lambda: False)

    service = LLMService()
    device, dtype = service._resolve_device()

    assert device == "cpu"
    assert dtype == torch.float32


def test_is_loaded_false_when_model_or_tokenizer_missing(add_llm_to_syspath):
    from llm import LLMService

    service = LLMService()
    assert service.is_loaded is False

    service.model = object()
    assert service.is_loaded is False

    service.model = None
    service.tokenizer = object()
    assert service.is_loaded is False


def test_is_loaded_true_when_both_present(add_llm_to_syspath):
    from llm import LLMService

    service = LLMService()
    service.model = object()
    service.tokenizer = object()

    assert service.is_loaded is True


def test_load_initializes_tokenizer_and_model_once(add_llm_to_syspath, monkeypatch):
    import llm as llm_module
    from llm import LLMService

    fake_tokenizer = FakeTokenizer()
    fake_model = FakeModel()

    tokenizer_calls = {"count": 0}
    model_calls = {"count": 0}

    def fake_tokenizer_loader(model_name, use_fast):
        tokenizer_calls["count"] += 1
        assert use_fast is True
        return fake_tokenizer

    def fake_model_loader(model_name, torch_dtype, low_cpu_mem_usage):
        model_calls["count"] += 1
        assert low_cpu_mem_usage is True
        return fake_model

    service = LLMService()

    monkeypatch.setattr(service, "_resolve_device", lambda: ("cpu", torch.float32))
    monkeypatch.setattr(llm_module.AutoTokenizer, "from_pretrained", fake_tokenizer_loader)
    monkeypatch.setattr(llm_module.AutoModelForCausalLM, "from_pretrained", fake_model_loader)

    service.load()
    service.load()

    assert tokenizer_calls["count"] == 1
    assert model_calls["count"] == 1
    assert service.tokenizer is fake_tokenizer
    assert service.model is fake_model
    assert fake_model.to_calls == ["cpu"]
    assert fake_model.eval_called is True


def test_generate_text_uses_chat_template_and_decodes_output(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    service = LLMService()
    service.tokenizer = FakeTokenizer()
    service.model = FakeModel()

    monkeypatch.setattr(service, "load", lambda: None)

    result = service._generate_text(
        system_prompt="SYSTEM_PROMPT",
        user_prompt="USER_PROMPT",
        max_new_tokens=123,
    )

    assert result == '{"ok": true}'

    chat_call = service.tokenizer.chat_calls[0]
    assert chat_call["messages"] == [
        {"role": "system", "content": "SYSTEM_PROMPT"},
        {"role": "user", "content": "USER_PROMPT"},
    ]
    assert chat_call["tokenize"] is False
    assert chat_call["add_generation_prompt"] is True

    tokenize_call = service.tokenizer.tokenize_calls[0]
    assert tokenize_call["text"] == "PROMPT_TEXT"
    assert tokenize_call["return_tensors"] == "pt"

    generate_call = service.model.generate_calls[0]
    assert generate_call["max_new_tokens"] == 123
    assert generate_call["do_sample"] is False
    assert generate_call["repetition_penalty"] == 1.03
    assert generate_call["eos_token_id"] == service.tokenizer.eos_token_id
    assert generate_call["pad_token_id"] == service.tokenizer.eos_token_id


def test_repair_json_returns_fixed_payload(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    service = LLMService()
    service.tokenizer = FakeTokenizer()
    service.model = FakeModel()

    monkeypatch.setattr(service, "load", lambda: None)

    result = service._repair_json(
        broken_response="broken",
        original_user_prompt="original prompt",
        system_prompt="system prompt",
        max_new_tokens=96,
    )

    assert result == {"ok": True}

    chat_call = service.tokenizer.chat_calls[0]
    messages = chat_call["messages"]
    assert len(messages) == 4
    assert messages[0]["role"] == "system"
    assert messages[1]["role"] == "user"
    assert messages[2]["role"] == "assistant"
    assert messages[3]["role"] == "user"


def test_summarize_returns_parsed_json_without_repair(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    service = LLMService()

    monkeypatch.setattr(
        service,
        "_generate_text",
        lambda system_prompt, user_prompt, max_new_tokens: '{"category":"macro","ok":true}',
    )

    repair_called = {"value": False}

    def fake_repair_json(**kwargs):
        repair_called["value"] = True
        return {"repaired": True}

    monkeypatch.setattr(service, "_repair_json", fake_repair_json)

    result = service.summarize(
        category="macro",
        news=["first news", "second news"],
        max_chars_per_news=100,
        max_new_tokens=128,
    )

    assert result == {"category": "macro", "ok": True}
    assert repair_called["value"] is False


def test_summarize_falls_back_to_repair_on_invalid_json(add_llm_to_syspath, monkeypatch):
    import llm as llm_module
    from llm import LLMService

    service = LLMService()

    monkeypatch.setattr(
        service,
        "_generate_text",
        lambda system_prompt, user_prompt, max_new_tokens: "definitely not json",
    )

    captured = {}

    def fake_repair_json(broken_response, original_user_prompt, system_prompt, max_new_tokens):
        captured["broken_response"] = broken_response
        captured["original_user_prompt"] = original_user_prompt
        captured["system_prompt"] = system_prompt
        captured["max_new_tokens"] = max_new_tokens
        return {"repaired": True}

    monkeypatch.setattr(service, "_repair_json", fake_repair_json)

    result = service.summarize(
        category="macro",
        news=["first news"],
        max_chars_per_news=50,
        max_new_tokens=128,
    )

    assert result == {"repaired": True}
    assert captured["broken_response"] == "definitely not json"
    assert "Category: macro" in captured["original_user_prompt"]
    assert "News (English):" in captured["original_user_prompt"]
    assert captured["system_prompt"] == llm_module.SUMMARIZE_SYSTEM_PROMPT
    assert captured["max_new_tokens"] == llm_module.settings.repair_max_new_tokens


def test_score_returns_parsed_json_without_repair(add_llm_to_syspath, monkeypatch):
    from llm import LLMService

    service = LLMService()

    monkeypatch.setattr(
        service,
        "_generate_text",
        lambda system_prompt, user_prompt, max_new_tokens: '{"category":"macro","score":0.1}',
    )

    repair_called = {"value": False}

    def fake_repair_json(**kwargs):
        repair_called["value"] = True
        return {"repaired": True}

    monkeypatch.setattr(service, "_repair_json", fake_repair_json)

    result = service.score(
        category="macro",
        summarization="Краткая сводка",
        features={"signal_direction": "up"},
        max_new_tokens=128,
    )

    assert result == {"category": "macro", "score": 0.1}
    assert repair_called["value"] is False


def test_score_falls_back_to_repair_on_invalid_json(add_llm_to_syspath, monkeypatch):
    import llm as llm_module
    from llm import LLMService

    service = LLMService()

    monkeypatch.setattr(
        service,
        "_generate_text",
        lambda system_prompt, user_prompt, max_new_tokens: "broken json",
    )

    captured = {}

    def fake_repair_json(broken_response, original_user_prompt, system_prompt, max_new_tokens):
        captured["broken_response"] = broken_response
        captured["original_user_prompt"] = original_user_prompt
        captured["system_prompt"] = system_prompt
        captured["max_new_tokens"] = max_new_tokens
        return {"repaired": True}

    monkeypatch.setattr(service, "_repair_json", fake_repair_json)

    result = service.score(
        category="macro",
        summarization="Краткая сводка",
        features={"signal_direction": "up", "signal_strength": 0.7},
        max_new_tokens=256,
    )

    assert result == {"repaired": True}
    assert captured["broken_response"] == "broken json"
    assert "Category: macro" in captured["original_user_prompt"]
    assert "Summarization (RU):" in captured["original_user_prompt"]
    assert "Features:" in captured["original_user_prompt"]
    assert captured["system_prompt"] == llm_module.SCORING_SYSTEM_PROMPT
    assert captured["max_new_tokens"] == llm_module.settings.repair_max_new_tokens
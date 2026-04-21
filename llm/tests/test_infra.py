from types import SimpleNamespace


def test_build_redis_client_parses_settings_correctly(infra_module, monkeypatch):
    captured = {}

    class FakeRedis:
        def __init__(self, **kwargs):
            captured.update(kwargs)

    fake_settings = SimpleNamespace(
        redis_addr="redis-host:6380",
        redis_llm_response_db=7,
        redis_password="secret-pass",
    )

    monkeypatch.setattr(infra_module, "Redis", FakeRedis)
    monkeypatch.setattr(infra_module, "settings", fake_settings)

    client = infra_module.build_redis_client()

    assert captured["host"] == "redis-host"
    assert captured["port"] == 6380
    assert captured["db"] == 7
    assert captured["password"] == "secret-pass"
    assert captured["decode_responses"] is True
    assert isinstance(client, FakeRedis)


def test_build_redis_client_uses_none_for_empty_password(infra_module, monkeypatch):
    captured = {}

    class FakeRedis:
        def __init__(self, **kwargs):
            captured.update(kwargs)

    fake_settings = SimpleNamespace(
        redis_addr="valkey:6379",
        redis_llm_response_db=0,
        redis_password="",
    )

    monkeypatch.setattr(infra_module, "Redis", FakeRedis)
    monkeypatch.setattr(infra_module, "settings", fake_settings)

    infra_module.build_redis_client()

    assert captured["host"] == "valkey"
    assert captured["port"] == 6379
    assert captured["db"] == 0
    assert captured["password"] is None
    assert captured["decode_responses"] is True


def test_build_kafka_producer_uses_expected_serializers(infra_module, monkeypatch):
    captured = {}

    class FakeKafkaProducer:
        def __init__(self, **kwargs):
            captured.update(kwargs)

    fake_settings = SimpleNamespace(
        kafka_brokers="kafka-broker:19092",
    )

    monkeypatch.setattr(infra_module, "KafkaProducer", FakeKafkaProducer)
    monkeypatch.setattr(infra_module, "settings", fake_settings)

    producer = infra_module.build_kafka_producer()

    assert captured["bootstrap_servers"] == ["kafka-broker:19092"]

    value_serializer = captured["value_serializer"]
    key_serializer = captured["key_serializer"]

    serialized_value = value_serializer({"category": "macro", "text": "привет"})
    assert isinstance(serialized_value, bytes)
    assert b'"category": "macro"' in serialized_value
    assert b'"text":' in serialized_value

    assert key_serializer("macro") == b"macro"
    assert key_serializer(None) is None
    assert isinstance(producer, FakeKafkaProducer)
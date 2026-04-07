import json
from redis import Redis
from kafka import KafkaProducer

from config import settings


def build_redis_client() -> Redis:
    host, port = settings.redis_addr.split(":")
    return Redis(
        host=host,
        port=int(port),
        db=settings.redis_llm_response_db,
        password=settings.redis_password or None,
        decode_responses=True,
    )


def build_kafka_producer() -> KafkaProducer:
    return KafkaProducer(
        bootstrap_servers=[settings.kafka_brokers],
        value_serializer=lambda v: json.dumps(v, ensure_ascii=False).encode("utf-8"),
        key_serializer=lambda k: k.encode("utf-8") if k is not None else None,
    )
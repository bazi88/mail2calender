import os
from typing import Optional
from dataclasses import dataclass


@dataclass
class RedisConfig:
    host: str = os.getenv("REDIS_HOST", "redis")
    port: int = int(os.getenv("REDIS_PORT", "6379"))
    password: Optional[str] = os.getenv("REDIS_PASSWORD")
    pool_size: int = int(os.getenv("REDIS_POOL_SIZE", "10"))
    retry_on_timeout: bool = True
    max_retries: int = 3
    socket_timeout: int = 5
    socket_connect_timeout: int = 5


@dataclass
class RateLimitConfig:
    requests_per_minute: int = int(os.getenv("RATE_LIMIT", "100"))
    burst_factor: float = float(os.getenv("RATE_LIMIT_BURST_FACTOR", "1.5"))
    window_size: int = 60  # seconds


@dataclass
class CacheConfig:
    ttl: int = int(os.getenv("CACHE_TTL", "3600"))  # 1 hour
    local_max_size: int = int(os.getenv("LOCAL_CACHE_SIZE", "1000"))
    version: str = os.getenv("CACHE_VERSION", "v1.0")


@dataclass
class TelemetryConfig:
    otel_endpoint: str = os.getenv("OTEL_OTLP_ENDPOINT", "otel-collector:4317")
    elasticsearch_host: str = os.getenv("ELASTICSEARCH_HOST", "elasticsearch:9200")
    elasticsearch_user: str = os.getenv("ELASTICSEARCH_USER", "elastic")
    elasticsearch_password: str = os.getenv("ELASTICSEARCH_PASSWORD", "")
    log_level: str = os.getenv("LOG_LEVEL", "INFO")


@dataclass
class ModelConfig:
    model_name: str = os.getenv("MODEL_NAME", "vinai/phobert-base")
    device: str = os.getenv("DEVICE", "cpu")
    batch_size: int = int(os.getenv("BATCH_SIZE", "16"))
    max_sequence_length: int = int(os.getenv("MAX_SEQUENCE_LENGTH", "512"))


@dataclass
class Config:
    redis: RedisConfig = RedisConfig()
    rate_limit: RateLimitConfig = RateLimitConfig()
    cache: CacheConfig = CacheConfig()
    telemetry: TelemetryConfig = TelemetryConfig()
    model: ModelConfig = ModelConfig()
    timezone: str = os.getenv("TIMEZONE", "Asia/Ho_Chi_Minh")

    def validate(self):
        """Validate configuration settings"""
        # Validate timezone
        import pytz

        if self.timezone not in pytz.all_timezones:
            raise ValueError(f"Invalid timezone: {self.timezone}")

        # Validate rate limits
        if self.rate_limit.requests_per_minute < 1:
            raise ValueError("Rate limit must be positive")

        # Validate Redis configuration
        if self.redis.pool_size < 1:
            raise ValueError("Redis pool size must be positive")

        # Validate cache configuration
        if self.cache.ttl < 0:
            raise ValueError("Cache TTL must be non-negative")

        # Validate model configuration
        if self.model.batch_size < 1:
            raise ValueError("Batch size must be positive")
        if self.model.max_sequence_length < 1:
            raise ValueError("Max sequence length must be positive")

        return self


# Global configuration instance
config = Config().validate()

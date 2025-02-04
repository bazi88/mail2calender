import asyncio
import time
from typing import Dict, Optional
import redis
from redis.retry import Retry
import grpc
from functools import wraps
import logging
from core.config import config
from utils.telemetry import telemetry

logger = logging.getLogger(__name__)


class RateLimiter:
    """Enhanced rate limiter implementation using Redis"""

    def __init__(self):
        retry_strategy = Retry(
            max_attempts=config.redis.max_retries,
            backoff=Retry.exponential_backoff(cap=10),
        )

        self.redis = redis.Redis(
            host=config.redis.host,
            port=config.redis.port,
            password=config.redis.password,
            decode_responses=True,
            retry=retry_strategy,
            retry_on_timeout=config.redis.retry_on_timeout,
            socket_timeout=config.redis.socket_timeout,
            socket_connect_timeout=config.redis.socket_connect_timeout,
            connection_pool=redis.ConnectionPool(
                max_connections=config.redis.pool_size,
                host=config.redis.host,
                port=config.redis.port,
                password=config.redis.password,
            ),
        )

        self.window = config.rate_limit.window_size
        self.rate_limit = config.rate_limit.requests_per_minute
        self.burst_limit = int(self.rate_limit * config.rate_limit.burst_factor)

    async def is_allowed(self, client_id: str) -> bool:
        """
        Check if request is allowed based on rate limit.

        Args:
            client_id: Unique identifier for the client (IP or API key)

        Returns:
            bool: True if request is allowed, False otherwise
        """
        try:
            with telemetry.create_span(
                "rate_limit_check",
                {"client_id": client_id, "rate_limit": self.rate_limit},
            ):
                pipe = self.redis.pipeline()
                now = int(time.time())
                key = f"rate_limit:{client_id}"

                # Clean old requests
                pipe.zremrangebyscore(key, 0, now - self.window)

                # Count requests in current window
                pipe.zcard(key)

                # Add current request
                pipe.zadd(key, {str(now): now})

                # Set expiry
                pipe.expire(key, self.window)

                # Execute pipeline
                _, request_count, *_ = pipe.execute()

                # Check against burst limit
                is_allowed = request_count <= self.burst_limit

                # Record metrics
                if not is_allowed:
                    telemetry.record_rate_limit()
                    logger.warning(
                        f"Rate limit exceeded for {client_id}: {request_count} requests"
                    )

                return is_allowed

        except redis.RedisError as e:
            logger.error(f"Redis error in rate limiter: {e}")
            # Log to Elasticsearch
            telemetry.log_to_elasticsearch(
                {"event": "rate_limit_error", "error": str(e), "client_id": client_id}
            )
            return True  # Allow request on Redis error to prevent service interruption

    def get_client_id(self, context: grpc.ServicerContext) -> str:
        """Extract client identifier from gRPC context"""
        peer = context.peer()
        if peer:
            return peer.split(":")[-1]  # Get IP address
        return "unknown"


def rate_limit(func):
    """Decorator to apply rate limiting to gRPC methods"""

    @wraps(func)
    async def wrapper(self, request, context, *args, **kwargs):
        limiter = getattr(self, "rate_limiter", None)
        if limiter is None:
            # Create rate limiter if not exists
            self.rate_limiter = RateLimiter()
            limiter = self.rate_limiter

        client_id = limiter.get_client_id(context)
        if not await limiter.is_allowed(client_id):
            error_msg = f"Rate limit exceeded. Please try again later."
            await context.abort(grpc.StatusCode.RESOURCE_EXHAUSTED, error_msg)
            # Log rate limit event
            telemetry.log_to_elasticsearch(
                {
                    "event": "rate_limit_exceeded",
                    "client_id": client_id,
                    "method": func.__name__,
                }
            )

        # Start timing the request
        start_time = time.time()
        try:
            result = await func(self, request, context, *args, **kwargs)
            duration = time.time() - start_time

            # Record metrics
            telemetry.record_processing_time(
                duration, {"method": func.__name__, "client_id": client_id}
            )

            return result

        except Exception as e:
            # Log error
            telemetry.log_to_elasticsearch(
                {
                    "event": "request_error",
                    "error": str(e),
                    "client_id": client_id,
                    "method": func.__name__,
                }
            )
            raise

    return wrapper

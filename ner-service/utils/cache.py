import json
import hashlib
import redis
from typing import Optional, Dict, Any
import logging
from datetime import timedelta

logger = logging.getLogger(__name__)


class Cache:
    """Redis-based cache implementation"""

    def __init__(
        self,
        redis_host: str = "redis",
        redis_port: int = 6379,
        redis_password: Optional[str] = None,
    ):
        self.redis = redis.Redis(
            host=redis_host,
            port=redis_port,
            password=redis_password,
            decode_responses=True,
        )
        self.default_ttl = timedelta(hours=1)
        self.version = "v1.0"  # Cache version for invalidation

    def _generate_key(self, text: str) -> str:
        """Generate cache key from input text"""
        text_hash = hashlib.md5(text.encode()).hexdigest()
        return f"ner_cache:{self.version}:{text_hash}"

    async def get(self, text: str) -> Optional[Dict[str, Any]]:
        """
        Get cached result for text

        Args:
            text: Input text to lookup

        Returns:
            Optional[Dict]: Cached result if exists, None otherwise
        """
        try:
            key = self._generate_key(text)
            cached = self.redis.get(key)
            if cached:
                logger.debug(f"Cache hit for key: {key}")
                return json.loads(cached)
            logger.debug(f"Cache miss for key: {key}")
            return None

        except (redis.RedisError, json.JSONDecodeError) as e:
            logger.error(f"Cache error: {e}")
            return None

    async def set(
        self, text: str, result: Dict[str, Any], ttl: Optional[timedelta] = None
    ) -> bool:
        """
        Cache result for text

        Args:
            text: Input text
            result: Result to cache
            ttl: Time-to-live (optional)

        Returns:
            bool: True if cached successfully, False otherwise
        """
        try:
            key = self._generate_key(text)
            ttl = ttl or self.default_ttl

            success = self.redis.setex(key, ttl, json.dumps(result))

            if success:
                logger.debug(f"Cached result for key: {key}")
            return bool(success)

        except (redis.RedisError, json.JSONDecodeError) as e:
            logger.error(f"Cache error: {e}")
            return False

    async def invalidate(self, text: str) -> bool:
        """
        Remove cached result for text

        Args:
            text: Input text

        Returns:
            bool: True if invalidated successfully, False otherwise
        """
        try:
            key = self._generate_key(text)
            return bool(self.redis.delete(key))

        except redis.RedisError as e:
            logger.error(f"Cache error: {e}")
            return False

    async def clear_all(self) -> bool:
        """
        Clear all cached results

        Returns:
            bool: True if cleared successfully, False otherwise
        """
        try:
            pattern = f"ner_cache:{self.version}:*"
            keys = self.redis.keys(pattern)
            if keys:
                return bool(self.redis.delete(*keys))
            return True

        except redis.RedisError as e:
            logger.error(f"Cache error: {e}")
            return False

    def update_version(self, new_version: str):
        """Update cache version to invalidate all existing entries"""
        self.version = new_version
        logger.info(f"Updated cache version to: {new_version}")

import logging
from typing import Dict, Any
import grpc
import os
from protos import ner_pb2
from protos import ner_pb2_grpc
from model.ner_model import NERModel
from utils.rate_limiter import rate_limit
from utils.cache import Cache
from utils.recurring_events import RecurringEventProcessor

logger = logging.getLogger(__name__)


class NERService(ner_pb2_grpc.NERServiceServicer):
    def __init__(self):
        self.model = NERModel()

        # Initialize cache
        self.cache = Cache(
            redis_host=os.getenv("REDIS_HOST", "redis"),
            redis_port=int(os.getenv("REDIS_PORT", 6379)),
            redis_password=os.getenv("REDIS_PASSWORD"),
        )

        # Initialize recurring event processor
        self.recurring_processor = RecurringEventProcessor()

        logger.info("NER Service initialized with cache and rate limiting")

    @rate_limit
    async def ExtractEntities(self, request, context):
        try:
            text = request.text
            if not text:
                await context.abort(
                    grpc.StatusCode.INVALID_ARGUMENT, "Text cannot be empty"
                )

            # Try to get from cache first
            cached_result = await self.cache.get(text)
            if cached_result:
                return self._create_response(cached_result)

            # Extract entities using model
            entities = self.model.extract_entities(text)

            # Process recurring events
            processed_entities = []
            for entity in entities:
                if entity["type"] in ["TIME", "DATE"]:
                    # Process recurring patterns
                    entity = self.recurring_processor.process_recurring_time(entity)
                processed_entities.append(entity)

            # Cache the result
            await self.cache.set(text, processed_entities)

            return self._create_response(processed_entities)

        except Exception as e:
            logger.error(f"Error in ExtractEntities: {str(e)}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Internal error: {str(e)}")

    @rate_limit
    async def BatchExtractEntities(self, request, context):
        try:
            texts = request.texts
            if not texts:
                await context.abort(
                    grpc.StatusCode.INVALID_ARGUMENT, "Texts list cannot be empty"
                )

            # Process batch using model
            batch_results = []
            for text in texts:
                # Try cache first
                cached_result = await self.cache.get(text)
                if cached_result:
                    batch_results.append(cached_result)
                    continue

                # Extract and process if not cached
                entities = self.model.extract_entities(text)

                # Process recurring events
                processed_entities = []
                for entity in entities:
                    if entity["type"] in ["TIME", "DATE"]:
                        entity = self.recurring_processor.process_recurring_time(entity)
                    processed_entities.append(entity)

                # Cache the result
                await self.cache.set(text, processed_entities)
                batch_results.append(processed_entities)

            # Convert results to proto format
            proto_results = []
            for entities in batch_results:
                response = self._create_response(entities)
                proto_results.append(response)

            return ner_pb2.BatchExtractEntitiesResponse(results=proto_results)

        except Exception as e:
            logger.error(f"Error in BatchExtractEntities: {str(e)}")
            await context.abort(grpc.StatusCode.INTERNAL, f"Internal error: {str(e)}")

    def _create_response(
        self, entities: list[Dict[str, Any]]
    ) -> ner_pb2.ExtractEntitiesResponse:
        """Convert entities to proto response format"""
        proto_entities = []
        for entity in entities:
            proto_entity = ner_pb2.Entity(
                text=entity["text"],
                type=entity["type"],
                confidence=float(entity["confidence"]),
                start_pos=entity["start_pos"],
                end_pos=entity["end_pos"],
            )

            # Add time-specific fields
            if "normalized_time" in entity:
                proto_entity.normalized_time = entity["normalized_time"]
            if "timestamp" in entity:
                proto_entity.timestamp = entity["timestamp"]

            # Add recurrence information if present
            if "recurrence" in entity:
                proto_entity.recurrence.type = entity["recurrence"]["type"]
                proto_entity.recurrence.rrule = entity["recurrence"]["rrule"]
                proto_entity.recurrence.next_occurrences.extend(
                    entity["recurrence"]["next_occurrences"]
                )

            proto_entities.append(proto_entity)

        return ner_pb2.ExtractEntitiesResponse(entities=proto_entities)

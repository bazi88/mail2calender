import grpc
from concurrent import futures
import logging
from typing import List
import time

from model.ner_model import NERModel
from utils.metrics import metrics
from config.config import config
import protos.ner_pb2 as ner_pb2
import protos.ner_pb2_grpc as ner_pb2_grpc

logger = logging.getLogger(__name__)

class NERService(ner_pb2_grpc.NERServiceServicer):
    def __init__(self):
        self.model = NERModel()
        logger.info(f"NER Service initialized with device: {config.DEVICE}")

    async def ExtractEntities(self, request, context):
        start_time = time.time()
        try:
            entities = self.model.extract_entities(request.text)

            process_time = time.time() - start_time
            metrics.processing_time.labels(method='extract_entities').observe(process_time)
            metrics.request_counter.labels(method='extract', status='success').inc()

            return ner_pb2.ExtractEntitiesResponse(entities=[
                ner_pb2.Entity(
                    text=e['text'],
                    type=e['type'],
                    confidence=e['confidence'],
                    start_pos=e['start_pos'],
                    end_pos=e['end_pos']
                )
                for e in entities
            ],
            process_time=process_time
            )
        except Exception as e:
            logger.error(f"Error extracting entities: {e}")
            metrics.request_counter.labels(method='extract', status= 'error').inc()
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ner_pb2.ExtractEntitiesResponse(error_message=str(e))
        
    async def BatchExtractEntities(self, request, context):
        start_time = time.time()
        try:
            batch_size = request.batch_size or config.DEFAULT_BATCH_SIZE
            metrics.batch_size.observe(len(request.requests))
            
            texts = [req.text for req in request.requests]
            all_entities = self.model.batch_extract_entities(texts, batch_size)
            
            process_time = time.time() - start_time
            metrics.processing_time.labels(method='batch_extract').observe(process_time)
            metrics.request_counter.labels(method='batch_extract', status='success').inc()
            
            responses = []
            for entities in all_entities:
                response = ner_pb2.NERResponse(
                    entities=[
                        ner_pb2.Entity(
                            text=e["text"],
                            type=e["type"],
                            confidence=e["confidence"],
                            start_pos=e["start_pos"],
                            end_pos=e["end_pos"]
                        ) for e in entities
                    ]
                )
                responses.append(response)
                
            return ner_pb2.BatchNERResponse(
                responses=responses,
                total_process_time=process_time
            )
            
        except Exception as e:
            logger.error(f"Error processing batch request: {e}")
            metrics.request_counter.labels(method='batch_extract', status='error').inc()
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ner_pb2.BatchNERResponse(error_message=str(e))
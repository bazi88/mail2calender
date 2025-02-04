# ner-service/main.py
import asyncio
import logging
import grpc
from concurrent import futures
from prometheus_client import start_http_server
from grpc_health.v1 import health_pb2_grpc

from server.grpc_server import NERService
from server.health_check import HealthServicer
from config.config import config
from protos import ner_pb2_grpc
from model.ner_model import NERModel

logger = logging.getLogger(__name__)

async def serve():
    # Start metrics server
    start_http_server(config.METRICS_PORT)
    logger.info(f"Metrics server started on port {config.METRICS_PORT}")
    
    # Initialize model
    model = NERModel()
    
    # Setup gRPC server
    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=[
            ('grpc.max_send_message_length', 50 * 1024 * 1024),
            ('grpc.max_receive_message_length', 50 * 1024 * 1024),
            ('grpc.keepalive_time_ms', 60000),
            ('grpc.keepalive_timeout_ms', 20000)
        ]
    )
    
    # Add services
    ner_servicer = NERService(model)
    health_servicer = HealthServicer(model)
    
    ner_pb2_grpc.add_NERServiceServicer_to_server(ner_servicer, server)
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    
    # Start server
    server.add_insecure_port(f'[::]:{config.GRPC_PORT}')
    await server.start()
    
    logger.info(f"NER Service started on port {config.GRPC_PORT}")
    
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down server...")
        await server.stop(0)

if __name__ == "__main__":
    logging.basicConfig(
        level=config.LOG_LEVEL,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    asyncio.run(serve())
import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
import logging
import torch

from model.ner_model import NERModel
from config.config import config

logger = logging.getLogger(__name__)

class HealthServicer(health_pb2_grpc.HealthServicer):
    def __init__(self, model: NERModel):
        self.model = model

    async def Check(self, request, context):
        try:
            # Kiểm tra model có load được không
            test_text = "Xin chào từ health check"
            _ = self.model.extract_entities(test_text)
            
            # Kiểm tra GPU memory nếu dùng CUDA
            if torch.cuda.is_available():
                gpu_memory = torch.cuda.memory_allocated()
                if gpu_memory > 0.95 * torch.cuda.get_device_properties(0).total_memory:
                    raise Exception("GPU memory nearly full")
            
            return health_pb2.HealthCheckResponse(
                status=health_pb2.HealthCheckResponse.SERVING
            )
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return health_pb2.HealthCheckResponse(
                status=health_pb2.HealthCheckResponse.NOT_SERVING
            )

    async def Watch(self, request, context):
        while True:
            response = await self.Check(request, context)
            yield response


import torch
import os
from dataclasses import dataclass

@dataclass
class Config:
    MODEL_NAME: str = "vinai/phobert-base"
    MAX_SEQUENCE_LENGTH: int = 512
    DEFAULT_BATCH_SIZE: int = 8
    CONFIDENCE_THRESHOLD: float = 0.5
    CACHE_SIZE: int = 1000
    GRPC_PORT: int = int(os.getenv('GRPC_PORT', '50051'))
    METRICS_PORT: int = int(os.getenv('METRICS_PORT', '8000'))
    LOG_LEVEL: str = os.getenv('LOG_LEVEL', 'INFO')
    DEVICE: str = os.getenv('DEVICE', 'cuda' if torch.cuda.is_available() else 'cpu')

config = Config()
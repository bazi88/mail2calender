from prometheus_client import Counter, Histogram, Gauge
import time

class MetricsCollector:
    def __init__(self):
        self.request_counter = Counter(
            'ner_requests_total',
            'Total number of NER requests',
            ['method', 'status']
        )
        
        self.processing_time = Histogram(
            'ner_processing_seconds',
            'Time spent processing requests',
            ['method']
        )
        
        self.batch_size = Histogram(
            'ner_batch_size',
            'Batch sizes of requests'
        )
        
        self.model_memory = Gauge(
            'ner_model_memory_bytes',
            'Memory used by NER model'
        )

metrics = MetricsCollector()
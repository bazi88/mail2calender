import logging
from opentelemetry import trace, metrics
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.grpc import GrpcInstrumentorServer
from opentelemetry.instrumentation.redis import RedisInstrumentor
from elasticsearch import Elasticsearch
import os

logger = logging.getLogger(__name__)


class Telemetry:
    """Centralized telemetry management for NER service"""

    def __init__(self):
        self._setup_tracing()
        self._setup_metrics()
        self._setup_logging()
        self._setup_elasticsearch()

    def _setup_tracing(self):
        """Configure OpenTelemetry tracing"""
        resource = Resource.create(
            {
                "service.name": "ner-service",
                "service.version": "1.0.0",
                "deployment.environment": os.getenv("ENVIRONMENT", "development"),
            }
        )

        tracer_provider = TracerProvider(resource=resource)
        otlp_exporter = OTLPSpanExporter(
            endpoint=os.getenv("OTEL_OTLP_ENDPOINT", "otel-collector:4317"),
            insecure=True,
        )
        tracer_provider.add_span_processor(otlp_exporter)
        trace.set_tracer_provider(tracer_provider)

        # Instrument gRPC server
        grpc_instrumentor = GrpcInstrumentorServer()
        grpc_instrumentor.instrument()

        # Instrument Redis
        redis_instrumentor = RedisInstrumentor()
        redis_instrumentor.instrument()

        self.tracer = trace.get_tracer(__name__)

    def _setup_metrics(self):
        """Configure OpenTelemetry metrics"""
        meter_provider = MeterProvider(
            resource=Resource.create({"service.name": "ner-service"})
        )
        metrics.set_meter_provider(meter_provider)

        otlp_exporter = OTLPMetricExporter(
            endpoint=os.getenv("OTEL_OTLP_ENDPOINT", "otel-collector:4317")
        )
        meter_provider.add_exporter(otlp_exporter)

        self.meter = metrics.get_meter(__name__)

        # Create metrics
        self.request_counter = self.meter.create_counter(
            "ner_requests_total", description="Total number of NER requests"
        )

        self.processing_time = self.meter.create_histogram(
            "ner_processing_time",
            description="Time taken to process NER requests",
            unit="s",
        )

        self.cache_hits = self.meter.create_counter(
            "ner_cache_hits_total", description="Total number of cache hits"
        )

        self.rate_limits = self.meter.create_counter(
            "ner_rate_limits_total", description="Total number of rate limit hits"
        )

    def _setup_logging(self):
        """Configure structured logging"""
        logging.basicConfig(
            level=os.getenv("LOG_LEVEL", "INFO"),
            format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        )

    def _setup_elasticsearch(self):
        """Configure Elasticsearch logging"""
        self.es = Elasticsearch(
            hosts=[os.getenv("ELASTICSEARCH_HOST", "elasticsearch:9200")],
            basic_auth=(
                os.getenv("ELASTICSEARCH_USER", "elastic"),
                os.getenv("ELASTICSEARCH_PASSWORD", ""),
            ),
        )

    def log_to_elasticsearch(self, log_data: dict):
        """Send log data to Elasticsearch"""
        try:
            self.es.index(
                index=f"ner-service-logs-{datetime.now():%Y.%m.%d}",
                document={
                    "timestamp": datetime.utcnow().isoformat(),
                    "service": "ner-service",
                    **log_data,
                },
            )
        except Exception as e:
            logger.error(f"Failed to log to Elasticsearch: {e}")

    def create_span(self, name: str, attributes: dict = None):
        """Create a new trace span"""
        return self.tracer.start_as_current_span(name, attributes=attributes or {})

    def record_request(self, method: str, status: str = "success"):
        """Record a request metric"""
        self.request_counter.add(1, {"method": method, "status": status})

    def record_processing_time(self, duration: float, labels: dict = None):
        """Record processing time metric"""
        self.processing_time.record(duration, labels or {})

    def record_cache_hit(self):
        """Record cache hit metric"""
        self.cache_hits.add(1)

    def record_rate_limit(self):
        """Record rate limit hit metric"""
        self.rate_limits.add(1)


# Global telemetry instance
telemetry = Telemetry()

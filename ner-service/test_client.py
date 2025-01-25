import asyncio
import grpc
import time
from protos import ner_pb2
from protos import ner_pb2_grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc

async def wait_for_service(max_retries=5, retry_delay=5):
    channel = grpc.aio.insecure_channel('ner-service:50051')
    health_stub = health_pb2_grpc.HealthStub(channel)
    
    for i in range(max_retries):
        try:
            response = await health_stub.Check(health_pb2.HealthCheckRequest())
            if response.status == health_pb2.HealthCheckResponse.SERVING:
                print("Service is ready!")
                await channel.close()
                return True
        except grpc.RpcError:
            print(f"Service not ready, retrying in {retry_delay} seconds...")
            await asyncio.sleep(retry_delay)
    
    await channel.close()
    return False

async def test_extract_entities():
    # Đợi service sẵn sàng
    if not await wait_for_service():
        print("Service không sẵn sàng sau nhiều lần thử")
        return
    
    # Tạo channel đến server
    channel = grpc.aio.insecure_channel('ner-service:50051')
    stub = ner_pb2_grpc.NERServiceStub(channel)
    
    # Tạo request
    request = ner_pb2.ExtractEntitiesRequest(
        text="Tôi có cuộc họp với anh Nam vào lúc 2 giờ chiều ngày mai tại văn phòng công ty ABC"
    )
    
    try:
        # Gọi service
        response = await stub.ExtractEntities(request)
        print("\nKết quả trích xuất:")
        print("-" * 50)
        for entity in response.entities:
            print(f"Text: {entity.text}")
            print(f"Type: {entity.type}")
            print(f"Confidence: {entity.confidence:.2f}")
            print(f"Position: {entity.start_pos} -> {entity.end_pos}")
            print("-" * 50)
        print(f"Thời gian xử lý: {response.process_time:.2f}s")
        
    except grpc.RpcError as e:
        print(f"Lỗi RPC: {e.details()}")
    
    await channel.close()

if __name__ == "__main__":
    asyncio.run(test_extract_entities()) 
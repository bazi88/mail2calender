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
    
    # Test cases cho nhiều ngôn ngữ
    test_cases = [
        # Tiếng Việt
        "Tôi có cuộc họp với anh Nam và chị Hương vào lúc 2 giờ chiều ngày mai tại văn phòng công ty ABC ở Hà Nội",
        "Bộ trưởng Nguyễn Văn A đã có chuyến thăm chính thức tới Microsoft tại Singapore vào tháng trước",
        "Trường Đại học Bách Khoa Hà Nội tổ chức hội thảo về AI tại Việt Nam",
        
        # Tiếng Anh
        "John Smith and Mary Johnson will meet with Google's CEO at their New York office tomorrow",
        "Apple announced their new iPhone at their headquarters in Cupertino, California",
        "The United Nations conference in Geneva discussed climate change with representatives from China and Russia",
        
        # Tiếng Trung
        "李明和王芳将在明天下午在北京微软公司与张总监会面讨论新项目",
        "中国科学院的研究人员在上海举办了一场关于人工智能的研讨会",
        "阿里巴巴集团在杭州总部宣布与腾讯合作新计划",
        
        # Tiếng Nhật
        "田中さんは明日東京のソニー本社で佐藤部長と山本社長と会議があります",
        "トヨタ自動車は名古屋工場で新型電気自動車の発表会を開催する",
        "日立製作所の鈴木部長は大阪支社の山田課長と打ち合わせを行う",
        
        # Tiếng Hàn
        "김영희는 내일 삼성전자 서울사무소에서 이부장과 박차장을 만날 예정입니다",
        "현대자동차는 울산공장에서 신형 전기차를 공개했습니다",
        "LG전자의 정회장은 부산지사의 최부장과 회의를 가졌습니다"
    ]
    
    for text in test_cases:
        try:
            # Gọi service
            request = ner_pb2.ExtractEntitiesRequest(text=text)
            response = await stub.ExtractEntities(request)
            
            print(f"\nInput text: {text}")
            print("\nKết quả trích xuất:")
            print("-" * 50)
            for entity in response.entities:
                print(f"Text: {entity.text}")
                print(f"Type: {entity.type}")
                print(f"Confidence: {entity.confidence:.2f}")
                print(f"Position: {entity.start_pos} -> {entity.end_pos}")
                print("-" * 50)
            print(f"Thời gian xử lý: {response.process_time:.2f}s")
            print("\n" + "=" * 80 + "\n")
            
        except grpc.RpcError as e:
            print(f"Lỗi RPC: {e.details()}")
            
    await channel.close()

if __name__ == "__main__":
    asyncio.run(test_extract_entities()) 
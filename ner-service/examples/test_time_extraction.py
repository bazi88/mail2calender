import asyncio
import grpc
from protos import ner_pb2, ner_pb2_grpc
from datetime import datetime
import pytz


async def test_time_extraction():
    """Test NER service with various time expressions"""

    # Create channel
    channel = grpc.aio.insecure_channel("localhost:50052")
    stub = ner_pb2_grpc.NERServiceStub(channel)

    # Test cases with different time formats and recurring patterns
    test_cases = [
        # Single time expressions
        "Họp lúc 14h30 ngày mai",
        "Cuộc họp diễn ra vào 9h sáng thứ 2 tuần sau",
        "Deadline dự án là 25/12/2023",
        # Recurring events
        "Họp team mỗi thứ 2 lúc 9h sáng",
        "Đi tập thể dục hàng ngày lúc 6h sáng",
        "Review code mỗi tuần vào thứ 6 lúc 15h",
        "Báo cáo KPI hàng tháng vào ngày 5",
        "Team building mỗi năm vào tháng 12",
    ]

    try:
        print("\n=== Testing Single Entity Extraction ===")
        for text in test_cases:
            # Test rate limiting and caching by making multiple requests
            for i in range(2):  # Second request should use cache
                response = await stub.ExtractEntities(
                    ner_pb2.ExtractEntitiesRequest(text=text)
                )

                print(f"\nInput ({i+1}): {text}")
                print("Entities found:")
                for entity in response.entities:
                    if entity.type in ["TIME", "DATE"]:
                        print(f"- Text: {entity.text}")
                        print(f"  Type: {entity.type}")
                        print(f"  Confidence: {entity.confidence:.2f}")
                        if entity.normalized_time:
                            print(f"  Normalized Time: {entity.normalized_time}")
                        if entity.timestamp:
                            dt = datetime.fromtimestamp(
                                entity.timestamp, pytz.timezone("Asia/Ho_Chi_Minh")
                            )
                            print(f"  Timestamp: {dt.isoformat()}")
                        if entity.HasField("recurrence"):
                            print("  Recurrence:")
                            print(f"    Type: {entity.recurrence.type}")
                            print(f"    RRULE: {entity.recurrence.rrule}")
                            print("    Next occurrences:")
                            for occ in entity.recurrence.next_occurrences[:3]:
                                print(f"      - {occ}")

        print("\n=== Testing Batch Processing ===")
        batch_response = await stub.BatchExtractEntities(
            ner_pb2.BatchExtractEntitiesRequest(texts=test_cases)
        )

        for i, result in enumerate(batch_response.results):
            print(f"\nBatch {i+1} - Input: {test_cases[i]}")
            for entity in result.entities:
                if entity.type in ["TIME", "DATE"]:
                    print(f"- Text: {entity.text}")
                    print(f"  Type: {entity.type}")
                    print(f"  Confidence: {entity.confidence:.2f}")
                    if entity.normalized_time:
                        print(f"  Normalized Time: {entity.normalized_time}")
                    if entity.HasField("recurrence"):
                        print("  Recurrence:")
                        print(f"    Type: {entity.recurrence.type}")
                        print(f"    RRULE: {entity.recurrence.rrule}")

        # Test rate limiting
        print("\n=== Testing Rate Limiting ===")
        for i in range(110):  # Should hit rate limit at 100
            try:
                response = await stub.ExtractEntities(
                    ner_pb2.ExtractEntitiesRequest(text="Test message")
                )
                print(f"Request {i+1}: Success")
            except grpc.RpcError as e:
                if e.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
                    print(f"Request {i+1}: Rate limit exceeded (expected)")
                    break
                else:
                    print(f"Request {i+1}: Error - {e.details()}")

    except Exception as e:
        print(f"Error: {str(e)}")
    finally:
        await channel.close()


if __name__ == "__main__":
    asyncio.run(test_time_extraction())

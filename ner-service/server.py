from concurrent import futures
import grpc
import ner_pb2
import ner_pb2_grpc

class NERService(ner_pb2_grpc.NERServiceServicer):
    def ExtractEntities(self, request, context):
        # Mock response với một số entity cố định
        entities = [
            ner_pb2.Entity(
                text="Apple",
                type="ORG",
                confidence=0.95,
                start_pos=0,
                end_pos=5
            ),
            ner_pb2.Entity(
                text="Cupertino",
                type="LOC",
                confidence=0.9,
                start_pos=20,
                end_pos=29
            ),
            ner_pb2.Entity(
                text="California",
                type="LOC",
                confidence=0.9,
                start_pos=31,
                end_pos=41
            )
        ]
        return ner_pb2.ExtractEntitiesResponse(entities=entities)

    def BatchExtractEntities(self, request, context):
        responses = []
        for req in request.requests:
            response = self.ExtractEntities(req, context)
            responses.append(response)
        return ner_pb2.BatchExtractEntitiesResponse(responses=responses)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ner_pb2_grpc.add_NERServiceServicer_to_server(NERService(), server)
    server.add_insecure_port('[::]:50051')
    print("Starting NER service on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve() 

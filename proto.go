package main

//go:generate protoc --proto_path=. --go_out=. --go_opt=module=mono-golang --go-grpc_out=. --go-grpc_opt=module=mono-golang ner-service/protos/ner.proto
//go:generate protoc --proto_path=. --go_out=. --go_opt=module=mono-golang --go-grpc_out=. --go-grpc_opt=module=mono-golang internal/domain/calendar/proto/calendar.proto

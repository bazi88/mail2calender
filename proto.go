//go:build ignore

package main

//go:generate protoc --proto_path=. --go_out=. --go_opt=module=mail2calendar --go-grpc_out=. --go-grpc_opt=module=mail2calendar ner-service/protos/ner.proto
//go:generate protoc --proto_path=. --go_out=. --go_opt=module=mail2calendar --go-grpc_out=. --go-grpc_opt=module=mail2calendar internal/domain/calendar/proto/calendar.proto

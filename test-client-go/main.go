package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	pb "test-client-go/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	address = "localhost:50051"
)

func waitForService(maxRetries int, retryDelay time.Duration) error {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		cancel()

		if err == nil && resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			fmt.Println("Service is ready!")
			return nil
		}

		fmt.Printf("Service not ready, retrying in %v seconds...\n", retryDelay.Seconds())
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("service did not become ready after %d retries", maxRetries)
}

func main() {
	// Wait for service to be ready
	if err := waitForService(5, 5*time.Second); err != nil {
		log.Fatalf("Failed to wait for service: %v", err)
	}

	// Set up a connection to the server
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create NER client
	client := pb.NewNERServiceClient(conn)

	// Test cases for multiple languages
	testCases := []string{
		// Vietnamese
		"Tôi có cuộc họp với anh Nam và chị Hương vào lúc 2 giờ chiều ngày mai tại văn phòng công ty ABC ở Hà Nội",
		"Bộ trưởng Nguyễn Văn A đã có chuyến thăm chính thức tới Microsoft tại Singapore vào tháng trước",
		"Trường Đại học Bách Khoa Hà Nội tổ chức hội thảo về AI tại Việt Nam",

		// English
		"John Smith and Mary Johnson will meet with Google's CEO at their New York office tomorrow",
		"Apple announced their new iPhone at their headquarters in Cupertino, California",
		"The United Nations conference in Geneva discussed climate change with representatives from China and Russia",

		// Chinese
		"李明和王芳将在明天下午在北京微软公司与张总监会面讨论新项目",
		"中国科学院的研究人员在上海举办了一场关于人工智能的研讨会",
		"阿里巴巴集团在杭州总部宣布与腾讯合作新计划",

		// Japanese
		"田中さんは明日東京のソニー本社で佐藤部長と山本社長と会議があります",
		"トヨタ自動車は名古屋工場で新型電気自動車の発表会を開催する",
		"日立製作所の鈴木部長は大阪支社の山田課長と打ち合わせを行う",

		// Korean
		"김영희는 내일 삼성전자 서울사무소에서 이부장과 박차장을 만날 예정입니다",
		"현대자동차는 울산공장에서 신형 전기차를 공개했습니다",
		"LG전자의 정회장은 부산지사의 최부장과 회의를 가졌습니다",
	}

	// Process each test case
	for _, text := range testCases {
		fmt.Printf("\nInput text: %s\n", text)

		// Create the request
		req := &pb.ExtractEntitiesRequest{Text: text}

		// Call the service
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.ExtractEntities(ctx, req)
		cancel()

		if err != nil {
			log.Printf("Error calling ExtractEntities: %v", err)
			continue
		}

		// Print results
		fmt.Println("\nExtracted entities:")
		fmt.Println(strings.Repeat("-", 50))
		for _, entity := range resp.Entities {
			fmt.Printf("Text: %s\n", entity.Text)
			fmt.Printf("Type: %s\n", entity.Type)
			fmt.Printf("Confidence: %.2f\n", entity.Confidence)
			fmt.Printf("Position: %d -> %d\n", entity.StartPos, entity.EndPos)
			fmt.Println(strings.Repeat("-", 50))
		}
		fmt.Printf("Processing time: %.2fs\n", resp.ProcessTime)
		fmt.Println(strings.Repeat("=", 80))
	}
}

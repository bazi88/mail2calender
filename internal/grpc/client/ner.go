package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mono-golang/internal/config"
	"mono-golang/internal/domain/ner"
	pb "mono-golang/ner-service/protos/ner"
)

// NER defines the interface for NER client operations
type NER interface {
	ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error)
}

type NERClient struct {
	client     pb.NERServiceClient
	connection *grpc.ClientConn
}

// NewNERClient creates a new NER service client
func NewNERClient(cfg *config.Config) (*NERClient, error) {
	address := fmt.Sprintf("%s:%d", cfg.NER.Host, cfg.NER.Port)

	// Create client using the recommended API
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NER service: %v", err)
	}

	client := pb.NewNERServiceClient(conn)
	return &NERClient{
		client:     client,
		connection: conn,
	}, nil
}

// ExtractEntities extracts named entities from text
func (c *NERClient) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	req := &pb.ExtractEntitiesRequest{
		Text: text,
	}
	resp, err := c.client.ExtractEntities(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %v", err)
	}

	// Convert protobuf response to domain response
	entities := make([]*ner.Entity, len(resp.Entities))
	for i, e := range resp.Entities {
		entities[i] = &ner.Entity{
			Text:  e.Text,
			Label: e.Type,
			Start: int(e.StartPos),
			End:   int(e.EndPos),
		}
	}

	return &ner.ExtractResponse{
		Entities: entities,
	}, nil
}

// Close closes the gRPC connection
func (c *NERClient) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// BatchExtractEntities extracts named entities from multiple texts
func (c *NERClient) BatchExtractEntities(ctx context.Context, requests []string, language string, batchSize int32) ([]*pb.ExtractEntitiesResponse, error) {
	batchReq := &pb.BatchExtractEntitiesRequest{
		BatchSize: batchSize,
		Requests:  make([]*pb.ExtractEntitiesRequest, len(requests)),
	}

	for i, text := range requests {
		batchReq.Requests[i] = &pb.ExtractEntitiesRequest{
			Text:     text,
			Language: language,
		}
	}

	resp, err := c.client.BatchExtractEntities(ctx, batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to batch extract entities: %v", err)
	}

	if resp.ErrorMessage != "" {
		return nil, fmt.Errorf("NER service error: %s", resp.ErrorMessage)
	}

	return resp.Responses, nil
}

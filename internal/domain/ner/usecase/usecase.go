package usecase

import (
	"context"
	"fmt"

	"mail2calendar/internal/domain/ner"
	"mail2calendar/internal/grpc/client"
)

// Entity represents a named entity in the usecase layer
type Entity struct {
	Text     string `json:"text"`
	Label    string `json:"label"`
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
}

type NER interface {
	ExtractEntities(ctx context.Context, text string) ([]*Entity, error)
}

type NERUseCase struct {
	client client.NER
}

// New creates a new NER use case
func New(client client.NER) *NERUseCase {
	return &NERUseCase{
		client: client,
	}
}

// ExtractEntities extracts named entities from the given text
func (uc *NERUseCase) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	return uc.client.ExtractEntities(ctx, text)
}

// ExtractEntitiesFromText extracts named entities from the given text and converts to internal format
func (uc *NERUseCase) ExtractEntitiesFromText(ctx context.Context, text string) ([]*ner.Entity, error) {
	response, err := uc.client.ExtractEntities(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %v", err)
	}

	if response == nil {
		return nil, nil
	}

	return response.Entities, nil
}

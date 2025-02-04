package ner

import "context"

// UseCase defines the interface for NER operations
type UseCase interface {
	// ExtractEntities extracts named entities from the given text
	ExtractEntities(ctx context.Context, text string) (*ExtractResponse, error)

	// ExtractEntitiesFromText extracts named entities from text and returns them in internal format
	ExtractEntitiesFromText(ctx context.Context, text string) ([]*Entity, error)
}

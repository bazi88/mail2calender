package ner

import "context"

// UseCase defines the interface for NER operations
type UseCase interface {
	// ExtractEntities extracts named entities from the given text
	ExtractEntities(ctx context.Context, text string) (*ExtractResponse, error)
}

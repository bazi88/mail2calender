package ner

// Entity represents a named entity extracted from text
type Entity struct {
	Text  string
	Label string
	Start int
	End   int
}

// ExtractResponse represents the response from entity extraction
type ExtractResponse struct {
	Entities []*Entity
}

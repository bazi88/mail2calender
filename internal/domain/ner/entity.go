package ner

// Entity represents a named entity extracted from text
type Entity struct {
	Text  string `json:"text"`
	Label string `json:"label"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// ExtractResponse represents the response from NER extraction
type ExtractResponse struct {
	Entities []*Entity `json:"entities"`
}

package repository

import (
	"context"
	"mono-golang/internal/domain/author"
)

// AuthorMock is a mock implementation of Author interface
type AuthorMock struct {
	CreateFunc func(ctx context.Context, author *author.CreateRequest) (*author.Schema, error)
	ReadFunc   func(ctx context.Context, authorID uint64) (*author.Schema, error)
	ListFunc   func(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error)
	UpdateFunc func(ctx context.Context, author *author.UpdateRequest) (*author.Schema, error)
	DeleteFunc func(ctx context.Context, authorID uint64) error
	SearchFunc func(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error)
}

// SearcherMock is a mock implementation of Searcher interface
type SearcherMock struct {
	SearchFunc func(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error)
}

func (m *AuthorMock) Create(ctx context.Context, author *author.CreateRequest) (*author.Schema, error) {
	return m.CreateFunc(ctx, author)
}

func (m *AuthorMock) Read(ctx context.Context, authorID uint64) (*author.Schema, error) {
	return m.ReadFunc(ctx, authorID)
}

func (m *AuthorMock) List(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error) {
	return m.ListFunc(ctx, filter)
}

func (m *AuthorMock) Update(ctx context.Context, author *author.UpdateRequest) (*author.Schema, error) {
	return m.UpdateFunc(ctx, author)
}

func (m *AuthorMock) Delete(ctx context.Context, authorID uint64) error {
	return m.DeleteFunc(ctx, authorID)
}

func (m *AuthorMock) Search(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error) {
	return m.SearchFunc(ctx, filter)
}

func (m *SearcherMock) Search(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error) {
	return m.SearchFunc(ctx, filter)
}

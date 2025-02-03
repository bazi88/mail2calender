package usecase

import (
	"context"
	"mono-golang/internal/domain/author"
)

// AuthorMock is a mock implementation of Author interface for handler tests
type AuthorMock struct {
	CreateFunc func(ctx context.Context, req *author.CreateRequest) (*author.Schema, error)
	ReadFunc   func(ctx context.Context, id uint64) (*author.Schema, error)
	ListFunc   func(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error)
	UpdateFunc func(ctx context.Context, req *author.UpdateRequest) (*author.Schema, error)
	DeleteFunc func(ctx context.Context, id uint64) error
	SearchFunc func(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error)
}

func (m *AuthorMock) Create(ctx context.Context, req *author.CreateRequest) (*author.Schema, error) {
	return m.CreateFunc(ctx, req)
}

func (m *AuthorMock) Read(ctx context.Context, id uint64) (*author.Schema, error) {
	return m.ReadFunc(ctx, id)
}

func (m *AuthorMock) List(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error) {
	return m.ListFunc(ctx, filter)
}

func (m *AuthorMock) Update(ctx context.Context, req *author.UpdateRequest) (*author.Schema, error) {
	return m.UpdateFunc(ctx, req)
}

func (m *AuthorMock) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFunc(ctx, id)
}

func (m *AuthorMock) Search(ctx context.Context, filter *author.Filter) ([]*author.Schema, int, error) {
	return m.SearchFunc(ctx, filter)
}

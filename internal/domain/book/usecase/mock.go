package usecase

import (
	"context"
	"mono-golang/internal/domain/book"
)

// BookMock is a mock implementation of Book interface for handler tests
type BookMock struct {
	CreateFunc func(ctx context.Context, req *book.CreateRequest) (*book.Schema, error)
	ReadFunc   func(ctx context.Context, id uint64) (*book.Schema, error)
	ListFunc   func(ctx context.Context, filter *book.Filter) ([]*book.Schema, error)
	UpdateFunc func(ctx context.Context, req *book.UpdateRequest) (*book.Schema, error)
	DeleteFunc func(ctx context.Context, id uint64) error
	SearchFunc func(ctx context.Context, filter *book.Filter) ([]*book.Schema, error)
}

func (m *BookMock) Create(ctx context.Context, req *book.CreateRequest) (*book.Schema, error) {
	return m.CreateFunc(ctx, req)
}

func (m *BookMock) Read(ctx context.Context, id uint64) (*book.Schema, error) {
	return m.ReadFunc(ctx, id)
}

func (m *BookMock) List(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	return m.ListFunc(ctx, filter)
}

func (m *BookMock) Update(ctx context.Context, req *book.UpdateRequest) (*book.Schema, error) {
	return m.UpdateFunc(ctx, req)
}

func (m *BookMock) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFunc(ctx, id)
}

func (m *BookMock) Search(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	return m.SearchFunc(ctx, filter)
}

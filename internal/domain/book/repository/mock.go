package repository

import (
	"context"
	"mono-golang/internal/domain/book"
)

// BookMock is a mock implementation of Book interface
type BookMock struct {
	CreateFunc func(ctx context.Context, book *book.CreateRequest) (uint64, error)
	ListFunc   func(ctx context.Context, filter *book.Filter) ([]*book.Schema, error)
	ReadFunc   func(ctx context.Context, bookID uint64) (*book.Schema, error)
	UpdateFunc func(ctx context.Context, book *book.UpdateRequest) error
	DeleteFunc func(ctx context.Context, bookID uint64) error
	SearchFunc func(ctx context.Context, filter *book.Filter) ([]*book.Schema, error)
}

func (m *BookMock) Create(ctx context.Context, book *book.CreateRequest) (uint64, error) {
	return m.CreateFunc(ctx, book)
}

func (m *BookMock) List(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	return m.ListFunc(ctx, filter)
}

func (m *BookMock) Read(ctx context.Context, bookID uint64) (*book.Schema, error) {
	return m.ReadFunc(ctx, bookID)
}

func (m *BookMock) Update(ctx context.Context, book *book.UpdateRequest) error {
	return m.UpdateFunc(ctx, book)
}

func (m *BookMock) Delete(ctx context.Context, bookID uint64) error {
	return m.DeleteFunc(ctx, bookID)
}

func (m *BookMock) Search(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	return m.SearchFunc(ctx, filter)
}

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mono-golang/internal/domain/book"
	"mono-golang/internal/domain/book/repository"
	"mono-golang/internal/utility/filter"
)

type mockBookRepo struct {
	mock.Mock
}

func (m *mockBookRepo) Create(ctx context.Context, book *book.Schema) error {
	args := m.Called(ctx, book)
	return args.Error(0)
}

func (m *mockBookRepo) List(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*book.Schema), args.Error(1)
}

func (m *mockBookRepo) GetByID(ctx context.Context, id uint64) (*book.Schema, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*book.Schema), args.Error(1)
}

func (m *mockBookRepo) Update(ctx context.Context, book *book.UpdateRequest) error {
	args := m.Called(ctx, book)
	return args.Error(0)
}

func (m *mockBookRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestBookUseCase_Create(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := NewBookUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		book := &book.Schema{
			Title:         "Test Book",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		repo.On("Create", ctx, book).Return(nil)

		err := useCase.Create(ctx, book)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Invalid Book", func(t *testing.T) {
		book := &book.Schema{} // Missing required fields

		err := useCase.Create(ctx, book)
		assert.Error(t, err)
	})

	t.Run("Repository Error", func(t *testing.T) {
		book := &book.Schema{
			Title:         "Test Book",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		expectedErr := errors.New("db error")
		repo.On("Create", ctx, book).Return(expectedErr)

		err := useCase.Create(ctx, book)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_List(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := NewBookUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		filter := &book.Filter{
			Base: filter.Filter{
				Page:          1,
				Offset:        0,
				Limit:         0,
				DisablePaging: false,
				Sort:          nil,
				Search:        false,
			},
			Title:         "Test",
			Description:   "",
			PublishedDate: "",
		}

		expectedBooks := []*book.Schema{
			{ID: 1, Title: "Test Book 1"},
			{ID: 2, Title: "Test Book 2"},
		}

		repo.On("List", ctx, filter).Return(expectedBooks, nil)

		books, err := useCase.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedBooks, books)
		repo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		filter := &book.Filter{}
		expectedErr := errors.New("db error")
		repo.On("List", ctx, filter).Return([]*book.Schema{}, expectedErr)

		books, err := useCase.List(ctx, filter)
		assert.Error(t, err)
		assert.Empty(t, books)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := NewBookUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		expectedBook := &book.Schema{
			ID:            1,
			Title:         "Test Book",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		repo.On("GetByID", ctx, uint64(1)).Return(expectedBook, nil)

		book, err := useCase.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedBook, book)
		repo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		repo.On("GetByID", ctx, uint64(999)).Return(nil, errors.New("not found"))

		book, err := useCase.GetByID(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, book)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_Update(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := NewBookUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		book := &book.UpdateRequest{
			Title:         "Updated Title",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Updated description",
		}

		repo.On("Update", ctx, book).Return(nil)

		err := useCase.Update(ctx, book)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Invalid Update", func(t *testing.T) {
		book := &book.UpdateRequest{} // Missing ID

		err := useCase.Update(ctx, book)
		assert.Error(t, err)
	})

	t.Run("Repository Error", func(t *testing.T) {
		book := &book.UpdateRequest{
			Title:         "Updated Title",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Updated description",
		}

		expectedErr := errors.New("db error")
		repo.On("Update", ctx, book).Return(expectedErr)

		err := useCase.Update(ctx, book)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := NewBookUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		repo.On("Delete", ctx, uint64(1)).Return(nil)

		err := useCase.Delete(ctx, 1)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo.On("Delete", ctx, uint64(999)).Return(expectedErr)

		err := useCase.Delete(ctx, 999)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_Search(t *testing.T) {
	type fields struct {
		bookRepo repository.Book
	}
	type args struct {
		ctx context.Context
		req *book.Filter
	}

	timeParsed, err := time.Parse(time.RFC3339, "2020-02-02T00:00:00Z")
	assert.Nil(t, err)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*book.Schema
		wantErr error
	}{
		{
			name: "simple",
			fields: fields{
				bookRepo: &repository.BookMock{
					SearchFunc: func(ctx context.Context, req *book.Filter) ([]*book.Schema, error) {
						return []*book.Schema{
							{
								ID:            1,
								Title:         "searched 1",
								PublishedDate: timeParsed,
								ImageURL:      "https://example.com/image1.png",
								Description:   "description",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: nil,
				req: &book.Filter{
					Base: filter.Filter{
						Page:          1,
						Offset:        0,
						Limit:         0,
						DisablePaging: false,
						Sort:          nil,
						Search:        true,
					},
					Title:         "searched",
					Description:   "",
					PublishedDate: "",
				},
			},
			want: []*book.Schema{
				{
					ID:            1,
					Title:         "searched 1",
					PublishedDate: timeParsed,
					ImageURL:      "https://example.com/image1.png",
					Description:   "description",
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &BookUseCase{
				bookRepo: tt.fields.bookRepo,
			}
			got, err := u.Search(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equalf(t, tt.want, got, "Search(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

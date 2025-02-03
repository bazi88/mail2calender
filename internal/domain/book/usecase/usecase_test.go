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
	"mono-golang/internal/utility/filter"
)

type mockBookRepo struct {
	mock.Mock
}

func (m *mockBookRepo) Create(ctx context.Context, req *book.CreateRequest) (uint64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockBookRepo) List(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*book.Schema), args.Error(1)
}

func (m *mockBookRepo) Read(ctx context.Context, id uint64) (*book.Schema, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*book.Schema), args.Error(1)
}

func (m *mockBookRepo) Update(ctx context.Context, req *book.UpdateRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockBookRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockBookRepo) Search(ctx context.Context, filter *book.Filter) ([]*book.Schema, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*book.Schema), args.Error(1)
}

func TestBookUseCase_Create(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

	t.Run("Success", func(t *testing.T) {
		repo.Mock = mock.Mock{} // Reset mock
		publishedDate := time.Now().Format("2006-01-02")
		req := &book.CreateRequest{
			Title:         "Test Book",
			PublishedDate: publishedDate,
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		expectedBook := &book.Schema{
			ID:            1,
			Title:         req.Title,
			PublishedDate: time.Now(),
			ImageURL:      req.ImageURL,
			Description:   req.Description,
		}

		repo.On("Create", ctx, req).Return(uint64(1), nil)
		repo.On("Read", ctx, uint64(1)).Return(expectedBook, nil)

		result, err := useCase.Create(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedBook, result)
		repo.AssertExpectations(t)
	})

	t.Run("Create Error", func(t *testing.T) {
		repo.Mock = mock.Mock{} // Reset mock
		publishedDate := time.Now().Format("2006-01-02")
		req := &book.CreateRequest{
			Title:         "Test Book",
			PublishedDate: publishedDate,
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		expectedErr := errors.New("db error")
		repo.On("Create", ctx, req).Return(uint64(0), expectedErr)
		// Don't expect Read call since Create returns error

		result, err := useCase.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_List(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

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
			{ID: 1, Title: "Test Book 1", PublishedDate: time.Now()},
			{ID: 2, Title: "Test Book 2", PublishedDate: time.Now()},
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

func TestBookUseCase_Read(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

	t.Run("Success", func(t *testing.T) {
		expectedBook := &book.Schema{
			ID:            1,
			Title:         "Test Book",
			PublishedDate: time.Now(),
			ImageURL:      "https://example.com/image.png",
			Description:   "Test description",
		}

		repo.On("Read", ctx, uint64(1)).Return(expectedBook, nil)

		book, err := useCase.Read(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedBook, book)
		repo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		repo.On("Read", ctx, uint64(999)).Return(nil, errors.New("not found"))

		book, err := useCase.Read(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, book)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_Update(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

	t.Run("Success", func(t *testing.T) {
		repo.Mock = mock.Mock{} // Reset mock
		publishedDate := time.Now().Format("2006-01-02")
		req := &book.UpdateRequest{
			ID:            1,
			Title:         "Updated Title",
			PublishedDate: publishedDate,
			ImageURL:      "https://example.com/image.png",
			Description:   "Updated description",
		}

		expectedBook := &book.Schema{
			ID:            req.ID,
			Title:         req.Title,
			PublishedDate: time.Now(),
			ImageURL:      req.ImageURL,
			Description:   req.Description,
		}

		repo.On("Update", ctx, req).Return(nil)
		repo.On("Read", ctx, uint64(1)).Return(expectedBook, nil)

		result, err := useCase.Update(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedBook, result)
		repo.AssertExpectations(t)
	})

	t.Run("Update Error", func(t *testing.T) {
		repo.Mock = mock.Mock{} // Reset mock
		publishedDate := time.Now().Format("2006-01-02")
		req := &book.UpdateRequest{
			ID:            1,
			Title:         "Updated Title",
			PublishedDate: publishedDate,
			ImageURL:      "https://example.com/image.png",
			Description:   "Updated description",
		}

		expectedErr := errors.New("db error")
		repo.On("Update", ctx, req).Return(expectedErr)
		// Don't expect Read call since Update returns error

		result, err := useCase.Update(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestBookUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

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
	ctx := context.Background()
	repo := new(mockBookRepo)
	useCase := New(repo)

	t.Run("Success", func(t *testing.T) {
		filter := &book.Filter{
			Base: filter.Filter{
				Search: true,
			},
			Title: "Test",
		}

		expectedBooks := []*book.Schema{
			{ID: 1, Title: "Test Book 1", PublishedDate: time.Now()},
			{ID: 2, Title: "Test Book 2", PublishedDate: time.Now()},
		}

		repo.On("Search", ctx, filter).Return(expectedBooks, nil)

		books, err := useCase.Search(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedBooks, books)
		repo.AssertExpectations(t)
	})

	t.Run("Search Error", func(t *testing.T) {
		filter := &book.Filter{
			Base: filter.Filter{
				Search: true,
			},
		}

		expectedErr := errors.New("search error")
		repo.On("Search", ctx, filter).Return([]*book.Schema{}, expectedErr)

		books, err := useCase.Search(ctx, filter)
		assert.Error(t, err)
		assert.Empty(t, books)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

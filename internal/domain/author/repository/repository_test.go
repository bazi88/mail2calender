package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewAuthorRepository()

	t.Run("Create Author", func(t *testing.T) {
		author := &Author{
			Name: "Test Author",
		}
		err := repo.Create(ctx, author)
		require.NoError(t, err)
		assert.NotZero(t, author.ID)
	})

	t.Run("Create Author with Books", func(t *testing.T) {
		author := &Author{
			Name: "Test Author with Books",
			Books: []Book{
				{Title: "Book 1"},
				{Title: "Book 2"},
			},
		}
		err := repo.Create(ctx, author)
		require.NoError(t, err)
		assert.NotZero(t, author.ID)
		assert.Len(t, author.Books, 2)
	})

	t.Run("Create Invalid Author", func(t *testing.T) {
		author := &Author{}
		err := repo.Create(ctx, author)
		require.Error(t, err)
	})
}

func TestAuthorRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewAuthorRepository()

	author := &Author{
		Name: "Original Name",
	}
	err := repo.Create(ctx, author)
	require.NoError(t, err)

	t.Run("Update Name", func(t *testing.T) {
		author.Name = "Updated Name"
		err := repo.Update(ctx, author)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, author.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("Update Non-existent Author", func(t *testing.T) {
		nonExistent := &Author{
			ID:   999,
			Name: "Non-existent",
		}
		err := repo.Update(ctx, nonExistent)
		require.Error(t, err)
	})
}

func TestAuthorRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewAuthorRepository()

	author := &Author{
		Name: "To Delete",
	}
	err := repo.Create(ctx, author)
	require.NoError(t, err)

	t.Run("Delete Existing", func(t *testing.T) {
		err := repo.Delete(ctx, author.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, author.ID)
		require.Error(t, err)
	})

	t.Run("Delete Non-existent", func(t *testing.T) {
		err := repo.Delete(ctx, 999)
		require.Error(t, err)
	})
}

func TestAuthorRepository_List(t *testing.T) {
	ctx := context.Background()
	repo := NewAuthorRepository()

	// Create test data
	authors := []*Author{
		{Name: "Author 1"},
		{Name: "Author 2"},
		{Name: "Author 3"},
	}
	for _, a := range authors {
		err := repo.Create(ctx, a)
		require.NoError(t, err)
	}

	t.Run("List All", func(t *testing.T) {
		result, err := repo.List(ctx, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("List with Filter", func(t *testing.T) {
		filter := &AuthorFilter{
			Name: "Author 1",
		}
		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Author 1", result[0].Name)
	})

	t.Run("List with Invalid Filter", func(t *testing.T) {
		filter := &AuthorFilter{
			Name: "Non-existent",
		}
		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

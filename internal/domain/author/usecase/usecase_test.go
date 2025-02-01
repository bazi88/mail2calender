package usecase

import (
	"context"
	"errors"
	"mono-golang/internal/domain/book"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"mono-golang/config"
	"mono-golang/internal/domain/author"
	"mono-golang/internal/domain/author/repository"
)

var c config.Cache

func TestMain(m *testing.M) {
	c = config.Cache{
		Enable: false,
	}
}

func TestAuthorUseCase_Create(t *testing.T) {
	type args struct {
		*author.CreateRequest
	}

	type want struct {
		Author *author.Schema
		err    error
	}

	type test struct {
		name string
		args
		want
	}

	tests := []test{
		{
			name: "simple",
			args: args{
				CreateRequest: &author.CreateRequest{
					FirstName:  "First",
					MiddleName: "Middle",
					LastName:   "Last",
					Books:      nil,
				},
			},
			want: want{
				Author: &author.Schema{
					ID:         1,
					FirstName:  "First",
					MiddleName: "Middle",
					LastName:   "Last",
					CreatedAt:  time.Time{},
					UpdatedAt:  time.Time{},
					DeletedAt:  nil,
					Books:      nil,
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoAuthor := &repository.AuthorMock{
				CreateFunc: func(ctx context.Context, r *author.CreateRequest) (*author.Schema, error) {
					return test.want.Author, test.want.err
				},
			}

			uc := New(c, repoAuthor, nil, nil, nil)

			got, err := uc.Create(context.Background(), test.args.CreateRequest)
			assert.Equal(t, test.want.err, err)
			assert.Equal(t, test.want.Author, got)
		})
	}
}

func TestAuthorUseCase_Update(t *testing.T) {
	type args struct {
		context.Context
		*author.UpdateRequest
	}
	type want struct {
		repo struct {
			*author.Schema
			error
		}
		error
	}

	type test struct {
		name string
		args
		want
	}

	createdTime := time.Now()

	tests := []test{
		{
			name: "simple",
			args: args{
				Context: context.Background(),
				UpdateRequest: &author.UpdateRequest{
					ID:         1,
					FirstName:  "Updated First",
					MiddleName: "Updated Middle",
					LastName:   "Updated Last",
				},
			},
			want: want{
				repo: struct {
					*author.Schema
					error
				}{
					&author.Schema{
						ID:         1,
						FirstName:  "Updated First",
						MiddleName: "Updated Middle",
						LastName:   "Updated Last",
						CreatedAt:  createdTime,
						UpdatedAt:  time.Now(),
						DeletedAt:  nil,
						Books:      []*book.Schema{},
					},
					nil,
				},
				error: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoAuthor := &repository.AuthorMock{
				UpdateFunc: func(ctx context.Context, authorParam *author.UpdateRequest) (*author.Schema, error) {
					return test.want.repo.Schema, test.want.repo.error
				},
			}

			cacheMock := &repository.AuthorRedisServiceMock{
				UpdateFunc: func(ctx context.Context, toAuthor *author.UpdateRequest) (*author.Schema, error) {
					return test.want.repo.Schema, test.want.repo.error
				},
			}

			uc := New(c, repoAuthor, nil, nil, cacheMock)

			update, err := uc.Update(test.args.Context, test.args.UpdateRequest)
			assert.Equal(t, test.want.error, err)

			assert.Equal(t, test.want.repo.ID, update.ID)
			assert.Equal(t, test.want.repo.FirstName, update.FirstName)
			assert.Equal(t, test.want.repo.MiddleName, update.MiddleName)
			assert.Equal(t, test.want.repo.LastName, update.LastName)
			assert.True(t, createdTime.Before(test.want.repo.CreatedAt) || createdTime.Equal(test.want.repo.CreatedAt))
			assert.True(t, createdTime.Before(test.want.repo.UpdatedAt) || createdTime.Equal(test.want.repo.UpdatedAt))
			assert.Nil(t, test.want.repo.DeletedAt)
		})
	}
}

func TestAuthorUseCase_Delete(t *testing.T) {
	type args struct {
		context.Context
		ID uint64
	}
	type want struct {
		error
	}
	type test struct {
		name string
		args
		want
	}

	tests := []test{
		{
			name: "simple",
			args: args{
				Context: context.Background(),
				ID:      1,
			},
			want: want{
				error: nil,
			},
		},
		{
			name: "zero ID",
			args: args{
				Context: context.Background(),
				ID:      0,
			},
			want: want{
				error: errors.New("ID cannot be 0 or less"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoAuthor := &repository.AuthorMock{
				DeleteFunc: func(ctx context.Context, authorID uint64) error {
					return test.want.error
				},
			}
			cacheMock := &repository.AuthorRedisServiceMock{
				DeleteFunc: func(ctx context.Context, id uint64) error {
					return test.want.error
				},
			}

			uc := New(c, repoAuthor, nil, nil, cacheMock)

			err := uc.Delete(test.args.Context, test.args.ID)
			assert.Equal(t, test.want.error, err)
		})
	}
}

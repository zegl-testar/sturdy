package graphql

//go:generate mockgen -destination internal/mock_db/user_repository_mock.go getsturdy.com/api/pkg/users/db Repository

import (
	"context"
	"testing"

	"getsturdy.com/api/pkg/author/graphql/internal/mock_db"
	gqldataloader "getsturdy.com/api/pkg/graphql/dataloader"
	"getsturdy.com/api/pkg/users"

	"github.com/golang/mock/gomock"
	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDataloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_db.NewMockRepository(ctrl)

	db.EXPECT().Get(gomock.Eq("user-id")).Return(&users.User{ID: "user-id", Name: "foo"}, nil).Times(1)
	db.EXPECT().Get(gomock.Eq("user-id2")).Return(&users.User{ID: "user-id2", Name: "foo"}, nil).Times(0)

	root := NewResolver(db, zap.NewNop())
	ctx := gqldataloader.NewContext(context.Background())

	// Get many times, expect to be cached, and the db to only be called once
	for i := 0; i < 5; i++ {
		author, err := root.Author(ctx, "user-id")
		assert.NoError(t, err)
		assert.Equal(t, graphql.ID("user-id"), author.ID())
	}
}

func TestDataloaderRequestedOncePerCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := mock_db.NewMockRepository(ctrl)

	db.EXPECT().Get(gomock.Eq("user-id")).Do(func(userID string) {
		t.Log("CALLED", userID)
	}).Return(&users.User{ID: "user-id", Name: "foo"}, nil).Times(5)

	root := NewResolver(db, zap.NewNop())

	// Get many times, expect to be cached, and the db to only be called once
	for i := 0; i < 5; i++ {
		ctx := gqldataloader.NewContext(context.Background())

		author, err := root.Author(ctx, "user-id")
		assert.NoError(t, err)
		assert.Equal(t, graphql.ID("user-id"), author.ID())

		author2, err := root.Author(ctx, "user-id")
		assert.NoError(t, err)
		assert.Equal(t, graphql.ID("user-id"), author2.ID())
	}
}

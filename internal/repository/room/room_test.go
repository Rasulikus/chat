package room

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Rasulikus/chat/internal/model"
	testdb "github.com/Rasulikus/chat/internal/repository/test_db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMain(m *testing.M) {
	testdb.RecreateTables()
	code := m.Run()
	testdb.CloseDB()
	os.Exit(code)
}

type testSuite struct {
	db       *bun.DB
	roomRepo *Repository
	ctx      context.Context
}

func setupTestSuite(t *testing.T) *testSuite {
	t.Helper()
	var suite testSuite
	suite.db = testdb.DB()
	suite.roomRepo = NewRepository(suite.db)
	suite.ctx = context.Background()
	return &suite
}

func Test_Repo_Insert(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)

	roomWithoutPassword := &model.Room{
		Name: "test room",
	}

	roomWithPassword := &model.Room{
		Name:         "testPassword room",
		PasswordHash: []byte("password"),
	}
	testCases := []struct {
		name    string
		room    *model.Room
		wantErr bool
	}{
		{
			name:    "insert without password",
			room:    roomWithoutPassword,
			wantErr: false,
		},
		{
			name:    "insert with password",
			room:    roomWithPassword,
			wantErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ts.roomRepo.Insert(ts.ctx, testCase.room)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, testCase.room.ID)
				assert.NotZero(t, testCase.room.Name)
				assert.NotZero(t, testCase.room.CreatedAt)
				assert.NotZero(t, testCase.room.UpdatedAt)
				assert.WithinDuration(t, time.Now(), testCase.room.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), testCase.room.UpdatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), testCase.room.LastActiveAt, time.Second)
				if testCase.room.ID == roomWithPassword.ID {
					assert.NotZero(t, testCase.room.PasswordHash)
				}
			}
		})
	}
}

func Test_Repo_GetByID(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)

	insertRoom := &model.Room{
		Name: "test room",
	}
	err := ts.roomRepo.Insert(ts.ctx, insertRoom)
	require.NoError(t, err)
	require.Equal(t, int64(1), insertRoom.ID)

	testCases := []struct {
		name    string
		roomID  int64
		wantErr bool
	}{
		{
			name:    "no err",
			roomID:  1,
			wantErr: false,
		},
		{
			name:    "not found err",
			roomID:  2,
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			room, err := ts.roomRepo.GetByID(ts.ctx, testCase.roomID)
			if testCase.wantErr {
				require.ErrorIs(t, err, model.ErrNotFound)
				require.Nil(t, room)
			} else {
				require.NoError(t, err)
				assert.Equal(t, insertRoom.Name, room.Name)
			}
		})
	}
}

func Test_Repo_List(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	insertRoom1 := &model.Room{
		Name: "test room",
	}
	insertRoom2 := &model.Room{
		Name: "test room 2",
	}
	err := ts.roomRepo.Insert(ts.ctx, insertRoom1)
	require.NoError(t, err)
	err = ts.roomRepo.Insert(ts.ctx, insertRoom2)
	require.NoError(t, err)

	t.Run("list all rooms", func(t *testing.T) {
		rooms, err := ts.roomRepo.List(ts.ctx, 10, "id ASC", nil)
		require.NoError(t, err)
		assert.Len(t, rooms, 2)
	})
	t.Run("list before 2 id rooms", func(t *testing.T) {
		rooms, err := ts.roomRepo.List(ts.ctx, 10, "id ASC", &insertRoom2.ID)
		require.NoError(t, err)
		assert.Len(t, rooms, 1)
		assert.Equal(t, insertRoom1.ID, rooms[0].ID)
	})
	t.Run("list with limit 1 rooms", func(t *testing.T) {
		rooms, err := ts.roomRepo.List(ts.ctx, 1, "id ASC", nil)
		require.NoError(t, err)
		assert.Len(t, rooms, 1)
		assert.Equal(t, insertRoom1.ID, rooms[0].ID)
	})
}

func Test_Repo_TouchActivity(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	insertRoom := &model.Room{
		Name:      "test room",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	err := ts.roomRepo.Insert(ts.ctx, insertRoom)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(-time.Hour), insertRoom.CreatedAt, time.Second)
	require.WithinDuration(t, time.Now().Add(-time.Hour), insertRoom.UpdatedAt, time.Second)

	testCases := []struct {
		name    string
		roomID  int64
		wantErr bool
	}{
		{"no err", insertRoom.ID, false},
		{"not found err", 9999999, true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ts.roomRepo.TouchActivity(ts.ctx, testCase.roomID)
			if testCase.wantErr {
				require.ErrorIs(t, err, model.ErrNotFound)
			} else {
				require.NoError(t, err)
				room, err := ts.roomRepo.GetByID(ts.ctx, testCase.roomID)
				require.NoError(t, err)
				assert.WithinDuration(t, time.Now(), room.UpdatedAt, time.Second)
			}
		})
	}
}

func Test_Repo_SoftDeleteInactiveOlderThan(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	insertRoom := &model.Room{
		Name:         "test room",
		LastActiveAt: time.Now().Add(-time.Hour),
	}
	err := ts.roomRepo.Insert(ts.ctx, insertRoom)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(-time.Hour), insertRoom.LastActiveAt, time.Second)

	t.Run("without delete", func(t *testing.T) {
		aff, err := ts.roomRepo.SoftDeleteInactiveOlderThan(ts.ctx, time.Minute*61)
		require.NoError(t, err)
		require.Zero(t, aff)
	})

	t.Run("soft delete inactive", func(t *testing.T) {
		aff, err := ts.roomRepo.SoftDeleteInactiveOlderThan(ts.ctx, time.Minute*59)
		require.NoError(t, err)
		require.Equal(t, int64(1), aff)
		_, err = ts.roomRepo.GetByID(ts.ctx, insertRoom.ID)
		require.ErrorIs(t, err, model.ErrNotFound)
	})
}

func Test_Repo_SoftDelete(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	insertRoom := &model.Room{
		Name: "test room",
	}
	err := ts.roomRepo.Insert(ts.ctx, insertRoom)
	require.NoError(t, err)
	t.Run("soft delete room", func(t *testing.T) {
		err = ts.roomRepo.SoftDelete(ts.ctx, insertRoom.ID)
		require.NoError(t, err)
		room, err := ts.roomRepo.GetByID(ts.ctx, insertRoom.ID)
		require.ErrorIs(t, err, model.ErrNotFound)
		require.Nil(t, room)
	})
}

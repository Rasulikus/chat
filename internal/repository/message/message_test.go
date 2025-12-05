package message

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/repository/room"
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
	db          *bun.DB
	messageRepo *Repository
	roomRepo    *room.Repository
	ctx         context.Context
}

func setupTestSuite(t *testing.T) *testSuite {
	t.Helper()
	var suite testSuite
	suite.db = testdb.DB()
	suite.messageRepo = NewRepository(suite.db)
	suite.roomRepo = room.NewRepository(suite.db)
	suite.ctx = context.Background()
	return &suite
}

func Test_Repo_Insert(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)

	testRoom := &model.Room{
		Name: "testroom",
	}
	err := ts.roomRepo.Insert(ts.ctx, testRoom)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		message *model.Message
		wantErr bool
	}{
		{
			name: "success",
			message: &model.Message{
				Nick:   "testNick",
				Text:   "some text",
				RoomID: testRoom.ID,
			},
			wantErr: false,
		},
		{
			name: "without not null value",
			message: &model.Message{
				Nick: "testNick",
				Text: "some text",
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err = ts.messageRepo.Insert(ts.ctx, testCase.message)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Zero(t, testCase.message.ID)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, testCase.message.ID)
				assert.NotZero(t, testCase.message.Nick)
				assert.NotZero(t, testCase.message.Text)
				assert.NotZero(t, testCase.message.CreatedAt)
				assert.WithinDuration(t, time.Now(), testCase.message.CreatedAt, time.Second)
			}
		})
	}
}

func Test_Repo_GetByID(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	testRoom := &model.Room{
		Name: "testroom",
	}
	err := ts.roomRepo.Insert(ts.ctx, testRoom)
	require.NoError(t, err)

	testMessage := &model.Message{
		Nick:   "testNick",
		Text:   "some text",
		RoomID: testRoom.ID,
	}
	err = ts.messageRepo.Insert(ts.ctx, testMessage)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "success",
			id:      testMessage.ID,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      -1,
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			message, err := ts.messageRepo.GetByID(ts.ctx, testCase.id)
			if testCase.wantErr {
				require.ErrorIs(t, err, model.ErrNotFound)
				assert.Nil(t, message)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, message.ID)
				assert.NotZero(t, message.Nick)
				assert.NotZero(t, message.Text)
				assert.NotZero(t, message.CreatedAt)
				assert.WithinDuration(t, time.Now(), message.CreatedAt, time.Second)
			}
		})
	}
}

func Test_Repo_ListByRoom(t *testing.T) {
	ts := setupTestSuite(t)
	testdb.CleanDB(ts.ctx)
	testRoom := &model.Room{
		Name: "testroom",
	}
	err := ts.roomRepo.Insert(ts.ctx, testRoom)
	require.NoError(t, err)
	testMessage1 := &model.Message{
		Nick:   "testNick",
		Text:   "some text",
		RoomID: testRoom.ID,
	}
	testMessage2 := &model.Message{
		Nick:   "testNick",
		Text:   "some other text",
		RoomID: testRoom.ID,
	}
	err = ts.messageRepo.Insert(ts.ctx, testMessage1)
	require.NoError(t, err)
	err = ts.messageRepo.Insert(ts.ctx, testMessage2)
	require.NoError(t, err)

	t.Run("list all messages in room", func(t *testing.T) {
		messages, err := ts.messageRepo.ListByRoom(ts.ctx, testRoom.ID, nil, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
	})

	t.Run("list before id 2 messages", func(t *testing.T) {
		messages, err := ts.messageRepo.ListByRoom(ts.ctx, testRoom.ID, &testMessage2.ID, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
	})

	t.Run("list with limit 1", func(t *testing.T) {
		messages, err := ts.messageRepo.ListByRoom(ts.ctx, testRoom.ID, nil, 1)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
	})

	t.Run("list with not valid room", func(t *testing.T) {
		messages, err := ts.messageRepo.ListByRoom(ts.ctx, -1, nil, 10)
		require.NoError(t, err)
		assert.Len(t, messages, 0)
	})
}

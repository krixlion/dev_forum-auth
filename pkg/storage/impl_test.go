package storage_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-auth/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	testCases := []struct {
		desc    string
		query   mocks.Query[entity.Token]
		args    args
		want    entity.Token
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want: entity.Token{},
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Token{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want:    entity.Token{},
			wantErr: true,
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Token{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mocks.Cmd[entity.Token]{}, tC.query, nulls.NullLogger{})
			got, err := db.Get(tC.args.ctx, tC.args.id)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Get():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}

			if !cmp.Equal(got, tC.want) {
				t.Errorf("storage.Get():\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
			assert.True(t, tC.query.AssertCalled(t, "Get", mock.Anything, tC.args.id))
		})
	}
}
func Test_GetMultiple(t *testing.T) {
	type args struct {
		ctx    context.Context
		offset string
		limit  string
	}

	testCases := []struct {
		desc    string
		query   mocks.Query[entity.Token]
		args    args
		want    []entity.Token
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want: []entity.Token{},
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Token{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want:    []entity.Token{},
			wantErr: true,
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Token{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mocks.Cmd[entity.Token]{}, tC.query, nulls.NullLogger{})
			got, err := db.GetMultiple(tC.args.ctx, tC.args.offset, tC.args.limit)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.GetMultiple():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}

			if !cmp.Equal(got, tC.want, cmpopts.EquateEmpty()) {
				t.Errorf("storage.GetMultiple():\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}

			assert.True(t, tC.query.AssertCalled(t, "GetMultiple", mock.Anything, tC.args.offset, tC.args.limit))
		})
	}
}
func Test_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		auth entity.Token
	}

	testCases := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Token]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:  context.Background(),
				auth: entity.Token{},
			},

			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if an error is forwarded",
			args: args{
				ctx:  context.Background(),
				auth: entity.Token{},
			},
			wantErr: true,
			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mocks.Query[entity.Token]{}, nulls.NullLogger{})
			err := db.Create(tC.args.ctx, tC.args.auth)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Create():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Create", mock.Anything, tC.args.auth))
		})
	}
}
func Test_Update(t *testing.T) {
	type args struct {
		ctx  context.Context
		auth entity.Token
	}

	testCases := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Token]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:  context.Background(),
				auth: entity.Token{},
			},

			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx:  context.Background(),
				auth: entity.Token{},
			},
			wantErr: true,
			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mocks.Query[entity.Token]{}, nulls.NullLogger{})
			err := db.Update(tC.args.ctx, tC.args.auth)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Update():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Update", mock.Anything, tC.args.auth))
		})
	}
}
func Test_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	testCases := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Token]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},

			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			wantErr: true,
			cmd: func() mocks.Cmd[entity.Token] {
				m := mocks.NewCmd[entity.Token]()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mocks.Query[entity.Token]{}, nulls.NullLogger{})
			err := db.Delete(tC.args.ctx, tC.args.id)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Delete():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Delete", mock.Anything, tC.args.id))
			assert.True(t, tC.cmd.AssertExpectations(t))
		})
	}
}

func Test_CatchUp(t *testing.T) {
	testCases := []struct {
		desc   string
		arg    event.Event
		query  mocks.Query[entity.Token]
		method string
	}{
		{
			desc: "Test if Update method is invoked on AuthUpdated event",
			arg: event.Event{
				Type: event.AuthUpdated,
				Body: gentest.RandomJSONauth(2, 3),
			},
			method: "Update",
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if Create method is invoked on AuthCreated event",
			arg: event.Event{
				Type: event.AuthCreated,
				Body: gentest.RandomJSONauth(2, 3),
			},
			method: "Create",
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Auth")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if Delete method is invoked on AuthDeleted event",
			arg: event.Event{
				Type: event.AuthDeleted,
				Body: func() []byte {
					id, err := json.Marshal(gentest.RandomString(5))
					if err != nil {
						t.Fatalf("Failed to marshal random ID to JSON. Error: %+v", err)
					}
					return id
				}(),
			},
			method: "Delete",
			query: func() mocks.Query[entity.Token] {
				m := mocks.NewQuery[entity.Token]()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mocks.Cmd[entity.Token]{}, tC.query, nulls.NullLogger{})
			db.CatchUp(tC.arg)

			switch tC.method {
			case "Delete":
				var id string
				err := json.Unmarshal(tC.arg.Body, &id)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON ID. Error: %+v", err)
					return
				}

				assert.True(t, tC.query.AssertCalled(t, tC.method, mock.Anything, id))

			default:
				var auth entity.Token
				err := json.Unmarshal(tC.arg.Body, &auth)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON auth. Error: %+v", err)
					return
				}

				assert.True(t, tC.query.AssertCalled(t, tC.method, mock.Anything, auth))
			}

			assert.True(t, tC.query.AssertExpectations(t))
		})
	}
}

package vault

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func Test_Make(t *testing.T) {
	t.Run("Test default logger and tracer are correctly assigned when not provided", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		got, err := Make(ctx, "host", "8888", "token", Config{MountPath: "path", KeyRefreshInterval: 0}, mocks.NewBroker(), nil, nil)
		if err != nil {
			t.Errorf("Make(): error = %v", err)
			return
		}

		if !cmp.Equal(got.logger, nulls.NullLogger{}) {
			t.Errorf("Make(): default logger not assigned:\n = %+v, want %+v", got.logger, nulls.NullLogger{})
		}

		if !cmp.Equal(got.tracer, nulls.NullTracer{}) {
			t.Errorf("Make(): default tracer not assigned:\n = %+v, want %+v", got.tracer, nulls.NullTracer{})
		}
	})

	t.Run("Test config is assigned", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		want := Config{
			MountPath:          "path",
			KeyCount:           50,
			KeyRefreshInterval: time.Hour,
		}

		got, err := Make(ctx, "host", "8888", "token", want, mocks.NewBroker(), nil, nil)
		if err != nil {
			t.Errorf("Make(): error = %v", err)
			return
		}

		if !cmp.Equal(got.config, want) {
			t.Errorf("Make():\n got = %v\n want = %v\n", got, want)
		}
	})

	t.Run("Test returns an error when given broker is nil", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		config := Config{
			MountPath:          "path",
			KeyCount:           50,
			KeyRefreshInterval: time.Hour,
		}

		if _, err := Make(ctx, "host", "8888", "token", config, nil, nil, nil); err == nil {
			t.Errorf("Make(): error = %v, wantErr = true", err)
			return
		}
	})
}

func TestConfig_validate(t *testing.T) {
	type fields struct {
		MountPath          string
		KeyCount           int
		KeyRefreshInterval time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test no error is returned on valid config",
			fields: fields{
				MountPath:          "/secret",
				KeyCount:           10,
				KeyRefreshInterval: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Test no error is returned on key refresh interval set to zero",
			fields: fields{
				MountPath:          "/secret",
				KeyCount:           10,
				KeyRefreshInterval: 0,
			},
			wantErr: false,
		},
		{
			name: "Test returns an error on key refresh interval set to any negative number",
			fields: fields{
				MountPath:          "/secret",
				KeyCount:           2,
				KeyRefreshInterval: -time.Minute,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				MountPath:          tt.fields.MountPath,
				KeyCount:           tt.fields.KeyCount,
				KeyRefreshInterval: tt.fields.KeyRefreshInterval,
			}
			if err := config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.validate():\n error = %v\n wantErr = %v\n", err, tt.wantErr)
			}
		})
	}
}

package vault

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func Test_Make(t *testing.T) {
	t.Run("Test default logger and tracer are correctly assigned when not provided", func(t *testing.T) {
		got, err := Make("host", "8888", "token", Config{
			MountPath: "path",
		}, nil, nil)
		if err != nil {
			t.Errorf("Make() error = %v", err)
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
		want := Config{
			MountPath:          "path",
			KeyCount:           50,
			KeyRefreshInterval: time.Hour,
		}

		got, err := Make("host", "8888", "token", want, nil, nil)
		if err != nil {
			t.Errorf("Make() error = %v", err)
			return
		}

		if !cmp.Equal(got.config, want) {
			t.Errorf("Make() = %v, want %v", got, want)
		}

	})
}

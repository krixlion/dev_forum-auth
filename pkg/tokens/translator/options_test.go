package translator

import (
	"testing"

	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func Test_optionFunc_apply(t *testing.T) {
	type args struct {
		t *Translator
	}
	tests := []struct {
		name string
		fn   optionFunc
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn.apply(tt.args.t)
		})
	}
}

func TestWithLogger(t *testing.T) {
	t.Run("Test given logger is assigned to translator", func(t *testing.T) {
		tr := &Translator{}
		logger := nulls.NullLogger{}
		optionFunc := WithLogger(logger)
		optionFunc.apply(tr)

		if tr.logger != logger {
			t.Errorf("WithLogger():\n got = %v\n want = %v", tr.logger, logger)
		}
	})
	t.Run("Test no-op when given logger is nil", func(t *testing.T) {
		tr := &Translator{logger: nulls.NullLogger{}}
		logger := (logging.Logger)(nil)
		optionFunc := WithLogger(logger)
		optionFunc.apply(tr)

		if tr.logger == logger {
			t.Errorf("WithLogger():\n got = %v\n want = %v", tr.logger, logger)
		}
	})
}

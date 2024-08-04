package translator

import (
	"testing"

	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/nulls"
	"go.opentelemetry.io/otel/trace"
)

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

func TestWithTracer(t *testing.T) {
	t.Run("Test given tracer is assigned to translator", func(t *testing.T) {
		tr := &Translator{}
		tracer := nulls.NullTracer{}
		optionFunc := WithTracer(tracer)
		optionFunc.apply(tr)

		if tr.tracer != tracer {
			t.Errorf("WithTracer():\n got = %v\n want = %v", tr.tracer, tracer)
		}
	})
	t.Run("Test no-op when given tracer is nil", func(t *testing.T) {
		tr := &Translator{tracer: nulls.NullTracer{}}
		tracer := (trace.Tracer)(nil)
		optionFunc := WithTracer(tracer)
		optionFunc.apply(tr)

		if tr.tracer == tracer {
			t.Errorf("WithTracer():\n got = %v\n want = %v", tr.tracer, tracer)
		}
	})
}

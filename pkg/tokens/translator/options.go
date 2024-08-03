package translator

import (
	"github.com/krixlion/dev_forum-lib/logging"
	"go.opentelemetry.io/otel/trace"
)

type Option interface {
	apply(*Translator)
}

type optionFunc func(*Translator)

func (fn optionFunc) apply(t *Translator) {
	fn(t)
}

// WithLogger sets the Translator's logger to a given logger.
// If given logger is nil then no logger is set and default settings apply.
func WithLogger(logger logging.Logger) Option {
	return optionFunc(func(t *Translator) {
		if logger != nil {
			t.logger = logger
		}
	})
}

// Withtracer sets the Translator's tracer to a given tracer.
// If given tracer is nil then no tracer is set and default settings apply.
func WithTracer(tracer trace.Tracer) Option {
	return optionFunc(func(t *Translator) {
		if tracer != nil {
			t.tracer = tracer
		}
	})
}

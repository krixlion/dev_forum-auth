package translator

import "github.com/krixlion/dev_forum-lib/logging"

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
	return optionFunc(func(st *Translator) {
		if logger != nil {
			st.logger = logger
		}
	})
}

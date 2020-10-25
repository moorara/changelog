package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name   string
		v      Verbosity
		format string
		args   []interface{}
	}{
		{
			name:   "Debug",
			v:      Debug,
			format: "foo: %s",
			args:   []interface{}{"bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := New(None)

			l.ChangeVerbosity(tc.v)

			l.Debug(tc.args...)
			l.Debugf(tc.format, tc.args...)

			l.Info(tc.args...)
			l.Infof(tc.format, tc.args...)

			l.Warn(tc.args...)
			l.Warnf(tc.format, tc.args...)

			l.Error(tc.args...)
			l.Errorf(tc.format, tc.args...)
		})
	}
}

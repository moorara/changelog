package log

import (
	"log"
	"os"
	"sync"

	"github.com/moorara/color"
)

// Verbosity is the verbosity level of a logger.
type Verbosity int

const (
	// Debug shows logs in all levels.
	Debug Verbosity = iota
	// Info shows logs in Info, Warn, Error, and Fatal levels.
	Info
	// Warn shows logs in Warn, Error, and Fatal levels.
	Warn
	// Error shows logs in Error and Fatal levels.
	Error
	// Fatal shows logs in Fatal level.
	Fatal
	// None does not show any logs.
	None
)

// Logger is a simple logger for logging to standard output.
type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

// logger implements the Logger interface.
type logger struct {
	sync.Mutex
	verbosity  Verbosity
	logger     *log.Logger
	debugColor *color.Color
	infoColor  *color.Color
	warnColor  *color.Color
	errorColor *color.Color
	fatalColor *color.Color
}

// New creates a new logger.
func New(v Verbosity) Logger {
	l := log.New(os.Stdout, "", log.LstdFlags)

	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	magenta := color.New(color.FgMagenta)
	red := color.New(color.FgRed)

	return &logger{
		verbosity:  v,
		logger:     l,
		debugColor: cyan,
		infoColor:  green,
		warnColor:  yellow,
		errorColor: magenta,
		fatalColor: red,
	}
}

func (l *logger) Debug(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Debug {
		msg := l.debugColor.Sprint(v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Debug {
		msg := l.debugColor.Sprintf(format, v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Info(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Info {
		msg := l.infoColor.Sprint(v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Info {
		msg := l.infoColor.Sprintf(format, v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Warn(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Warn {
		msg := l.warnColor.Sprint(v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Warn {
		msg := l.warnColor.Sprintf(format, v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Error(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Error {
		msg := l.errorColor.Sprint(v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Error {
		msg := l.errorColor.Sprintf(format, v...)
		l.logger.Print(msg)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Fatal {
		msg := l.fatalColor.Sprint(v...)
		l.logger.Fatal(msg)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if l.verbosity <= Fatal {
		msg := l.fatalColor.Sprintf(format, v...)
		l.logger.Fatal(msg)
	}
}

package logger

import (
	"log/slog"
	"os"

	"github.com/go-logr/logr/slogr"

	controllerlog "sigs.k8s.io/controller-runtime/pkg/log"
)

type Options struct {
	Development bool
}

var L *slog.Logger

func Init(opts Options) {
	ch := ContextHandler{}

	if opts.Development {
		ch.Handler = slog.Handler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}))
	} else {
		ch.Handler = slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelInfo,
			ReplaceAttr: nil,
		}))
	}

	l := slog.New(ch)
	slog.SetDefault(l)

	L = l

	controllerlog.SetLogger(slogr.NewLogr(ch.Handler))
}

func With(args ...any) *slog.Logger {
	return L.With(args...)
}

func WithGroup(name string) *slog.Logger {
	return L.WithGroup(name)
}

package logger

import (
	"log/slog"
	"os"

	"github.com/go-logr/logr/slogr"

	controllerlog "sigs.k8s.io/controller-runtime/pkg/log"
)

var L *slog.Logger

func Init(development bool) {
	ch := ContextHandler{}

	if development {
		ch.Handler = slog.Handler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
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

	L = slog.New(ch)
	slog.SetDefault(L)
	controllerlog.SetLogger(slogr.NewLogr(L.Handler()))
}

func With(args ...any) *slog.Logger {
	return L.With(args...)
}

func WithGroup(name string) *slog.Logger {
	return L.WithGroup(name)
}

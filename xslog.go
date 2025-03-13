package xslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type LogConfig struct {
	LogToConsole bool
	LogToFile    bool
	LogFilePath  string
}

type Logger struct {
	consoleLogger *slog.Logger
	fileLogger    *slog.Logger
	config        LogConfig
}

type TxtColoredHandler struct {
	out io.Writer
	mu  sync.Mutex
}

func NewTxtColoredHandler(out io.Writer) *TxtColoredHandler {
	return &TxtColoredHandler{out: out}
}

func (h *TxtColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *TxtColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	levelStr := fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", getLevelColor(r.Level), getLevelName(r))
	//levelStr := fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", levelColor, strings.ToUpper(r.Level.String()))

	//timeStr := r.Time.Format("2006-01-02 15:04:05")
	//msg := fmt.Sprintf("%s [%s] %s", timeStr, levelStr, r.Message)
	msg := fmt.Sprintf("[%s] %s", levelStr, r.Message)

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		return true
	})

	if len(attrs) > 0 {
		msg += " " + strings.Join(attrs, " ")
	}

	_, err := fmt.Fprintln(h.out, msg)
	return err
}

func (h *TxtColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TxtColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

func getLevelColor(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		return 35 // Purple
	case slog.LevelInfo:
		return 34 // Blue
	case slog.LevelWarn:
		return 33 // Yellow
	case slog.LevelError:
		return 31 // Red
	default:
		return 37 // Default White
	}
}

func getLevelName(r slog.Record) string {
	switch r.Level {
	case slog.LevelDebug:
		return "DBG"
	case slog.LevelInfo:
		return "INF"
	case slog.LevelWarn:
		return "WRN"
	case slog.LevelError:
		return "ERR"
	default:
		return strings.ToUpper(r.Level.String()[:3])
	}
}

func NewLogger(config LogConfig) (*Logger, error) {
	ml := &Logger{config: config}

	if config.LogToConsole {
		ml.consoleLogger = slog.New(NewTxtColoredHandler(os.Stdout))
	}

	if config.LogToFile {
		dir, _ := filepath.Split(config.LogFilePath)
		if len(dir) > 0 {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		}
		file, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		ml.fileLogger = slog.New(slog.NewJSONHandler(file, nil))
	}

	return ml, nil
}

func (ml *Logger) Info(msg string, args ...any) {
	if ml.config.LogToConsole {
		ml.consoleLogger.Info(msg, args...)
	}
	if ml.config.LogToFile {
		ml.fileLogger.Info(msg, args...)
	}
}

func (ml *Logger) Warn(msg string, args ...any) {
	if ml.config.LogToConsole {
		ml.consoleLogger.Warn(msg, args...)
	}
	if ml.config.LogToFile {
		ml.fileLogger.Warn(msg, args...)
	}
}

func (ml *Logger) Error(msg string, args ...any) {
	if ml.config.LogToConsole {
		ml.consoleLogger.Error(msg, args...)
	}
	if ml.config.LogToFile {
		ml.fileLogger.Error(msg, args...)
	}
}

func (ml *Logger) Debug(msg string, args ...any) {
	if ml.config.LogToConsole {
		ml.consoleLogger.Debug(msg, args...)
	}
	if ml.config.LogToFile {
		ml.fileLogger.Debug(msg, args...)
	}
}

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
	LogToConsole    bool
	LogToFile       bool
	LogFilePath     string
	LevelForFile    slog.Level
	LevelForConsole slog.Level
}

type Logger struct {
	consoleLogger   *slog.Logger
	fileLogger      *slog.Logger
	config          LogConfig
	consoleLevelVar *slog.LevelVar // 用于动态控制控制台日志级别
	fileLevelVar    *slog.LevelVar // 用于动态控制文件日志级别
	fileWriter      io.Writer      // 保存文件写入器，方便后续操作
}

type TxtColoredHandler struct {
	out  io.Writer
	opts *slog.HandlerOptions
	mu   sync.Mutex
}

func NewTxtColoredHandler(out io.Writer, opts *slog.HandlerOptions) *TxtColoredHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &TxtColoredHandler{
		opts: opts,
		out:  out,
	}
}

func (h *TxtColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// 如果没有设置 Level，则默认启用所有级别
	if h.opts.Level == nil {
		return true
	}
	// 检查当前级别是否符合配置的级别
	return level >= h.opts.Level.Level()
}

func (h *TxtColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	levelStr := fmt.Sprintf("\x1b[%dm%s\x1b[0m", getLevelColor(r.Level), getLevelName(r))
	//levelStr := fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", levelColor, strings.ToUpper(r.Level.String()))

	//timeStr := r.Time.Format("2006-01-02 15:04:05")
	//msg := fmt.Sprintf("%s [%s] %s", timeStr, levelStr, r.Message)
	msg := fmt.Sprintf("[%s] %s", levelStr, r.Message)

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%v", a.Value.Any()))
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
	ml := &Logger{
		config:          config,
		consoleLevelVar: new(slog.LevelVar),
		fileLevelVar:    new(slog.LevelVar),
	}

	// 设置初始级别
	ml.consoleLevelVar.Set(config.LevelForConsole)
	ml.fileLevelVar.Set(config.LevelForFile)

	if config.LogToConsole {
		ml.consoleLogger = slog.New(NewTxtColoredHandler(os.Stdout, &slog.HandlerOptions{
			Level: ml.consoleLevelVar,
		}))
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
		ml.fileWriter = file // 保存文件写入器
		ml.fileLogger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: ml.fileLevelVar,
		}))
	}

	return ml, nil
}

// 设置控制台日志级别
func (ml *Logger) SetConsoleLevel(level slog.Level) {
	if ml.consoleLevelVar != nil {
		ml.consoleLevelVar.Set(level)
	}
}

// 设置文件日志级别
func (ml *Logger) SetFileLevel(level slog.Level) {
	if ml.fileLevelVar != nil {
		ml.fileLevelVar.Set(level)
	}
}

// 获取控制台当前日志级别
func (ml *Logger) GetConsoleLevel() slog.Level {
	if ml.consoleLevelVar != nil {
		return ml.consoleLevelVar.Level()
	}
	return slog.LevelInfo // 默认
}

// 获取文件当前日志级别
func (ml *Logger) GetFileLevel() slog.Level {
	if ml.fileLevelVar != nil {
		return ml.fileLevelVar.Level()
	}
	return slog.LevelInfo // 默认
}

// 启用/禁用控制台日志
func (ml *Logger) EnableConsole(enable bool) {
	if enable && !ml.config.LogToConsole && ml.consoleLogger == nil {
		ml.consoleLevelVar = new(slog.LevelVar)
		ml.consoleLevelVar.Set(ml.config.LevelForConsole)
		ml.consoleLogger = slog.New(NewTxtColoredHandler(os.Stdout, &slog.HandlerOptions{
			Level: ml.consoleLevelVar,
		}))
	}
	ml.config.LogToConsole = enable
}

// 启用/禁用文件日志
func (ml *Logger) EnableFile(enable bool) error {
	// 如果要禁用且当前已启用
	if !enable && ml.config.LogToFile {
		ml.config.LogToFile = false
		// 如果有文件句柄，可以考虑关闭它
		if closer, ok := ml.fileWriter.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}

	// 如果要启用且当前未启用
	if enable && !ml.config.LogToFile && ml.fileLogger == nil {
		dir, _ := filepath.Split(ml.config.LogFilePath)
		if len(dir) > 0 {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		file, err := os.OpenFile(ml.config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		ml.fileWriter = file
		ml.fileLevelVar = new(slog.LevelVar)
		ml.fileLevelVar.Set(ml.config.LevelForFile)
		ml.fileLogger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: ml.fileLevelVar,
		}))
		ml.config.LogToFile = true
	}

	return nil
}

// ChangeFilePath 更改文件路径
func (ml *Logger) ChangeFilePath(newPath string) error {
	// 如果新路径与当前路径相同，无需操作
	if ml.config.LogFilePath == newPath {
		return nil
	}

	// 如果文件日志未启用，只更新配置
	if !ml.config.LogToFile {
		ml.config.LogFilePath = newPath
		return nil
	}

	// 关闭现有的文件
	if closer, ok := ml.fileWriter.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close existing log file: %w", err)
		}
	}

	// 创建新文件
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for new log file: %w", err)
	}

	file, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	// 更新配置和日志器
	ml.config.LogFilePath = newPath
	ml.fileWriter = file
	ml.fileLogger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: ml.fileLevelVar,
	}))

	return nil
}

// 关闭日志器，清理资源
func (ml *Logger) Close() error {
	if closer, ok := ml.fileWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (ml *Logger) Info(msg string, args ...any) {
	if ml.config.LogToConsole && ml.consoleLogger != nil {
		ml.consoleLogger.Info(msg, args...)
	}
	if ml.config.LogToFile && ml.fileLogger != nil {
		ml.fileLogger.Info(msg, args...)
	}
}

func (ml *Logger) Warn(msg string, args ...any) {
	if ml.config.LogToConsole && ml.consoleLogger != nil {
		ml.consoleLogger.Warn(msg, args...)
	}
	if ml.config.LogToFile && ml.fileLogger != nil {
		ml.fileLogger.Warn(msg, args...)
	}
}

func (ml *Logger) Error(msg string, args ...any) {
	if ml.config.LogToConsole && ml.consoleLogger != nil {
		ml.consoleLogger.Error(msg, args...)
	}
	if ml.config.LogToFile && ml.fileLogger != nil {
		ml.fileLogger.Error(msg, args...)
	}
}

func (ml *Logger) Debug(msg string, args ...any) {
	if ml.config.LogToConsole && ml.consoleLogger != nil {
		ml.consoleLogger.Debug(msg, args...)
	}
	if ml.config.LogToFile && ml.fileLogger != nil {
		ml.fileLogger.Debug(msg, args...)
	}
}

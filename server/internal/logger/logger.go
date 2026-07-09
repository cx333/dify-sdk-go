// Package logger 提供基于 slog 的结构化日志封装。
// 支持 JSON/Text 双格式输出、日志级别动态控制、日志文件轮转，以及 context 注入 request-id。
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 封装 slog.Logger，对外提供带 context 的便捷方法。
type Logger struct {
	inner *slog.Logger
}

// Config 日志初始化配置。
type Config struct {
	Level  string      // "debug" | "info" | "warn" | "error"，默认 info
	Format string      // "json" | "text"，开发环境默认 text
	Output io.Writer   // 输出目标，为 nil 时默认 os.Stdout
	File   *FileConfig // 文件轮转配置，为 nil 时仅输出到 Output
}

// FileConfig 日志文件轮转配置（基于 lumberjack）。
type FileConfig struct {
	Filename   string // 日志文件完整路径，如 "./logs/server.log"
	MaxSize    int    // 单个文件最大大小（MB），默认 100
	MaxAge     int    // 保留天数，默认 30
	MaxBackups int    // 最多保留的旧文件数，默认 10
	Compress   bool   // 是否压缩旧文件
}

// DefaultConfig 返回开发环境默认配置（仅输出到 stdout）。
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "text",
		Output: os.Stdout,
	}
}

// New 根据配置创建 Logger。
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}

	level := parseLevel(cfg.Level)
	writer := buildWriter(cfg)
	handler := newHandler(writer, level, cfg.Format)

	return &Logger{inner: slog.New(handler)}
}

// NewNop 返回丢弃所有输出的 Logger，用于测试。
func NewNop() *Logger {
	return &Logger{
		inner: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// Inner 返回底层 *slog.Logger，供需要直接使用的场景（如 Fiber 中间件集成）。
func (l *Logger) Inner() *slog.Logger {
	return l.inner
}

// ----- 带 context 的日志方法 -----

// Debug 输出 debug 级别日志，自动携带 context 中的 request-id。
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelDebug, msg, args...)
}

// Info 输出 info 级别日志。
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelInfo, msg, args...)
}

// Warn 输出 warn 级别日志。
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelWarn, msg, args...)
}

// Error 输出 error 级别日志。
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
}

// Fatal 输出 error 级别日志后调用 os.Exit(1)。
func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.log(ctx, slog.LevelError, msg, args...)
	os.Exit(1)
}

// ----- 无 context 的格式化日志 -----

// Debugf 无 context 的 debug 日志。
func (l *Logger) Debugf(format string, args ...any) {
	l.log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Infof 无 context 的 info 日志。
func (l *Logger) Infof(format string, args ...any) {
	l.log(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Warnf 无 context 的 warn 日志。
func (l *Logger) Warnf(format string, args ...any) {
	l.log(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Errorf 无 context 的 error 日志。
func (l *Logger) Errorf(format string, args ...any) {
	l.log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

// Fatalf 无 context 的 fatal 日志。
func (l *Logger) Fatalf(format string, args ...any) {
	l.log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// ----- context key & helper -----

type ctxKey string

const requestIDKey ctxKey = "request_id"

// WithRequestID 将 request-id 注入 context，后续日志自动携带。
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom 从 context 提取 request-id。
func RequestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// ----- 内部方法 -----

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if rid := RequestIDFrom(ctx); rid != "" {
		prefixed := make([]any, 0, len(args)+2)
		prefixed = append(prefixed, "request_id", rid)
		prefixed = append(prefixed, args...)
		l.inner.Log(ctx, level, msg, prefixed...)
	} else {
		l.inner.Log(ctx, level, msg, args...)
	}
}

func buildWriter(cfg Config) io.Writer {
	if cfg.File == nil {
		return cfg.Output
	}

	// 自动拼接文件名
	filename := cfg.File.Filename
	if filename == "" {
		filename = filepath.Join("logs", "server.log")
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    cfg.File.MaxSize,
		MaxAge:     cfg.File.MaxAge,
		MaxBackups: cfg.File.MaxBackups,
		Compress:   cfg.File.Compress,
		LocalTime:  true,
	}

	// 同时输出到控制台和文件
	return io.MultiWriter(cfg.Output, fileWriter)
}

func newHandler(w io.Writer, level slog.Level, format string) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // 由 callerHandler 自行计算 source，避免深度错位
	}

	var h slog.Handler
	switch strings.ToLower(format) {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	default:
		h = slog.NewTextHandler(w, opts)
	}

	return &callerHandler{handler: h}
}

// callerHandler 自行计算 source，跳过 logger 和 slog 包帧，定位到业务调用方。
type callerHandler struct {
	handler slog.Handler
}

func (h *callerHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *callerHandler) Handle(ctx context.Context, r slog.Record) error {
	f, line := findCaller()
	r.AddAttrs(slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f, line)))
	return h.handler.Handle(ctx, r)
}

func (h *callerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &callerHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *callerHandler) WithGroup(name string) slog.Handler {
	return &callerHandler{handler: h.handler.WithGroup(name)}
}

// findCaller 向上遍历调用栈，跳过 logger 和 slog 包，返回业务调用方的文件（短路径）和行号。
func findCaller() (file string, line int) {
	for skip := 3; ; skip++ {
		_, f, l, ok := runtime.Caller(skip)
		if !ok {
			return "???", 0
		}
		if !strings.Contains(f, "log/slog") &&
			!strings.Contains(f, "internal/logger/logger.go") {
			if idx := strings.Index(f, "/server/"); idx != -1 {
				f = f[idx+len("/server/"):]
			}
			return f, l
		}
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

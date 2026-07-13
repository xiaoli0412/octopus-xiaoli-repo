package log

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger
var atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
var consoleEncoder = zapcore.EncoderConfig{
	TimeKey:       "time",
	LevelKey:      "level",
	MessageKey:    "msg",
	CallerKey:     "caller",
	StacktraceKey: "stacktrace",
	EncodeLevel:   zapcore.CapitalLevelEncoder,
	EncodeTime:    zapcore.RFC3339TimeEncoder,
	EncodeCaller:  zapcore.ShortCallerEncoder,
}

func init() {
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoder),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	}
	Logger = zap.New(core, opts...).Sugar()
}

// Configure re-initializes the logger to also write to a rotating file when
// file is non-empty. When file is empty the logger keeps writing to stdout
// only (backward compatible). When file is set, logs are written to both the
// rotating file (via lumberjack) and stderr.
func Configure(file string, maxSize, maxBackups, maxAge int) {
	if file == "" {
		return
	}
	if dir := filepath.Dir(file); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	lw := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   false,
	}
	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoder),
		zapcore.AddSync(lw),
		atomicLevel,
	)
	stderrCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoder),
		zapcore.AddSync(os.Stderr),
		atomicLevel,
	)
	core := zapcore.NewTee(fileCore, stderrCore)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	}
	Logger = zap.New(core, opts...).Sugar()
}

func SetLevel(level string) {
	var lvl zapcore.Level
	err := lvl.UnmarshalText([]byte(level))
	if err != nil {
		return
	}
	atomicLevel.SetLevel(lvl)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

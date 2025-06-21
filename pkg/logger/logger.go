package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/nekruzjm/glb/pkg/config"
)

type Logger interface {
	Debug(log string, fields ...zapcore.Field)
	Info(log string, fields ...zapcore.Field)
	Warning(log string, fields ...zapcore.Field)
	Error(log string, fields ...zapcore.Field)
	Flush()
}

type logger struct {
	logger *zap.SugaredLogger
}

func New(cfg config.Config) Logger {
	var (
		stdoutSyncer = zapcore.Lock(os.Stdout)
		level        zapcore.Level
	)

	switch cfg.GetString("logger.level") {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.DebugLevel
	}

	prodEncoderConfig := zap.NewProductionEncoderConfig()
	prodEncoderConfig.FunctionKey = "func"

	core := zapcore.NewTee(zapcore.NewCore(
		zapcore.NewJSONEncoder(prodEncoderConfig), stdoutSyncer, level),
		zapcore.NewCore(getEncoder(cfg), getWriter(cfg), level),
	)

	sugaredLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

	return &logger{
		logger: sugaredLogger,
	}
}

func getEncoder(cfg config.Config) zapcore.Encoder {
	var encoderCfg = zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "time",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	if cfg.GetString("logger.env") == "local" {
		return zapcore.NewConsoleEncoder(encoderCfg)
	}

	return zapcore.NewJSONEncoder(encoderCfg)
}

func getWriter(cfg config.Config) zapcore.WriteSyncer {
	var log = &lumberjack.Logger{
		Filename:   cfg.GetString("logger.filename"),
		MaxSize:    cfg.GetInt("logger.maxSize"),
		MaxBackups: cfg.GetInt("logger.maxBackups"),
		MaxAge:     cfg.GetInt("logger.maxAge"),
		Compress:   false,
	}

	if log.Filename == "" {
		log.Filename = "./app.log"
	}
	if log.MaxSize == 0 {
		log.MaxSize = 200
	}
	if log.MaxBackups == 0 {
		log.MaxBackups = 10
	}
	if log.MaxAge == 0 {
		log.MaxAge = 30
	}

	return zapcore.AddSync(log)
}

func (l *logger) Debug(log string, fields ...zapcore.Field) {
	l.logger.Desugar().Debug(log, fields...)
}

func (l *logger) Info(log string, fields ...zapcore.Field) {
	l.logger.Desugar().Info(log, fields...)
}

func (l *logger) Warning(log string, fields ...zapcore.Field) {
	l.logger.Desugar().Warn(log, fields...)
}

func (l *logger) Error(log string, fields ...zapcore.Field) {
	l.logger.Desugar().Error(log, fields...)
}

func (l *logger) Flush() {
	_ = l.logger.Desugar().Sync()
}

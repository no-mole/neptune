package logger

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func init() {
	instance, _ := zap.NewProduction()
	SetLogger(NewLogger(context.Background(), instance))
}

var defaultLogger Logger

// SetLogger set global logger
func SetLogger(l Logger) {
	defaultLogger = l
}

// SetLoggerBridge set a warped logger for opentelemetry
func SetLoggerBridge() {
	// anyway, global.GetLoggerProvider will return a provider.
	provider := global.GetLoggerProvider()

	if provider != nil {
		core := zapcore.NewTee(
			zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), zap.InfoLevel),
			otelzap.NewCore("neptune/logger", otelzap.WithLoggerProvider(provider)),
		)
		instance := zap.New(core)
		SetLogger(NewLogger(context.Background(), instance))
	}
}

func NewLogger(ctx context.Context, l *zap.Logger) Logger {
	instance := &logger{
		ctx:      ctx,
		level:    LevelDebug,
		handlers: make([]Handle, 0),
		instance: l,
	}
	go func() {
		<-ctx.Done()
		_ = l.Sync()
	}()
	return instance
}

func Fatal(ctx context.Context, msg string, err error, fields ...zap.Field) {
	defaultLogger.Fatal(ctx, msg, err, fields...)
}

func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	defaultLogger.Error(ctx, msg, err, fields...)
}

func Warning(ctx context.Context, msg string, err error, fields ...zap.Field) {
	defaultLogger.Warning(ctx, msg, err, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Info(ctx, msg, fields...)
}

func Notice(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Notice(ctx, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Debug(ctx, msg, fields...)
}

func Trace(ctx context.Context, msg string, fields ...zap.Field) {
	defaultLogger.Trace(ctx, msg, fields...)
}

func AddHandle(ctx context.Context, handle Handle) {
	defaultLogger.Handle(ctx, handle)
}

type logger struct {
	//ctx when ctx.Done(),flush logs
	ctx context.Context

	//level only log < Level.Code()'s Entry
	level Level

	//handlers append field to entry
	handlers []Handle

	instance *zap.Logger
}

func (l *logger) Handle(_ context.Context, handle Handle) {
	l.handlers = append(l.handlers, handle)
}

func (l *logger) SetLevel(level Level) {
	l.level = level
}

func (l *logger) Fatal(ctx context.Context, msg string, err error, fields ...zap.Field) {
	l.logger(ctx, LevelFatal, msg, err, fields...)
}

func (l *logger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	l.logger(ctx, LevelError, msg, err, fields...)
}

func (l *logger) Warning(ctx context.Context, msg string, err error, fields ...zap.Field) {
	l.logger(ctx, LevelWarn, msg, err, fields...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger(ctx, LevelInfo, msg, nil, fields...)
}

func (l *logger) Notice(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger(ctx, LevelNotice, msg, nil, fields...)
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger(ctx, LevelDebug, msg, nil, fields...)
}

func (l *logger) Trace(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger(ctx, LevelTrace, msg, nil, fields...)
}

func (l *logger) logger(ctx context.Context, curLevel Level, msg string, err error, fields ...zap.Field) {
	if curLevel.Code() > l.level.Code() {
		return
	}
	span := trace.SpanFromContext(ctx)

	if span.IsRecording() {
		if span.SpanContext().HasTraceID() {
			fields = append(fields, WithField("trace_id", span.SpanContext().TraceID()))
		}
		if span.SpanContext().HasSpanID() {
			fields = append(fields, WithField("span_id", span.SpanContext().SpanID()))
		}
	}

	if err != nil {
		fields = append(fields, zap.NamedError("errorMsg", err))
	}
	for _, handle := range l.handlers {
		fields = append(fields, handle(ctx, curLevel, msg, err, fields...)...)
	}
	switch curLevel.Code() {
	case LevelTrace.Code(), LevelDebug.Code():
		l.instance.Debug(msg, fields...)
	case LevelInfo.Code():
		l.instance.Info(msg, fields...)
	case LevelWarn.Code():
		l.instance.Warn(msg, fields...)
	case LevelError.Code(), LevelFatal.Code():
		l.instance.Error(msg, fields...)
	}

	if span.IsRecording() {
		attributes := make([]attribute.KeyValue, 0, 2+len(fields))
		attributes = append(attributes, attribute.String("level", curLevel.String()))
		attributes = append(attributes, attribute.String("msg", msg))

		encode := zapcore.NewMapObjectEncoder()
		for _, field := range fields {
			field.AddTo(encode)
		}
		for k, v := range encode.Fields {
			attributes = append(attributes, attribute.String(k, fmt.Sprintf("%v", v)))
		}
		span.AddEvent("log", trace.WithAttributes(attributes...))
	}
}

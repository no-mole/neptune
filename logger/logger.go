package logger

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/no-mole/neptune/logger/dispatcher"
	"github.com/no-mole/neptune/logger/entry"
	"github.com/no-mole/neptune/logger/formatter"
	"github.com/no-mole/neptune/logger/level"
	"github.com/no-mole/neptune/logger/tagger"
	"gopkg.in/yaml.v3"
)

type BootstrapConfig struct {
	Level       string `json:"level" yaml:"level"`
	QueueLength int    `yaml:"queueLength"`
	Dispatchers []struct {
		Type      string            `json:"type" yaml:"type"`
		Formatter string            `json:"formatter" yaml:"formatter"`
		Tags      []string          `json:"tags" yaml:"tags"`
		Settings  map[string]string `json:"settings" yaml:"settings"`
	} `json:"dispatchers" yaml:"dispatchers"`
}

// Bootstrap init logger config
func Bootstrap(ctx context.Context, configBody []byte) error {
	if defaultLogger != nil {
		//停止默认logger，重新boot
		defaultLogger.Stop()
	}
	config := &BootstrapConfig{
		QueueLength: 1024,
	}
	err := yaml.Unmarshal(configBody, config)
	if err != nil {
		return err
	}
	l := NewLogger(ctx, config.QueueLength)

	l.SetLevel(level.GetLevelByName(config.Level))

	for _, item := range config.Dispatchers {
		formatterInstance, err := formatter.NewFormatter(item.Formatter)
		if err != nil {
			return err
		}

		dispatcherInstance, err := dispatcher.NewDispatcher(
			item.Type,
			formatterInstance,
			tagger.NewMatcher(item.Tags),
			&dispatcher.Config{Fields: item.Settings},
		)
		if err != nil {
			return err
		}

		l.AddDispatcher(dispatcherInstance)
	}
	return nil
}

var once sync.Once

func initDefaultLogger() {
	once.Do(func() {
		if defaultLogger == nil {
			//默认输出到stdout
			defaultLogger = NewLogger(context.Background(), 1024)
			defaultLogger.AddDispatcher(&dispatcher.StdoutDispatcher{Helper: dispatcher.NewHelper(&formatter.JsonFormatter{}, tagger.NewMatcher(nil))})
		}
	})
}

var defaultLogger Logger

// SetLogger set global logger
func SetLogger(l Logger) {
	defaultLogger = l
}

// GetLogger get global logger
func GetLogger() Logger {
	return defaultLogger
}

func NewLogger(ctx context.Context, queueLength int) Logger {
	instance := &logger{
		ctx:         ctx,
		level:       level.LevelDebug,
		logsCh:      make(chan entry.Entry, queueLength),
		flushCh:     make(chan struct{}, 1),
		dispatchers: make([]dispatcher.Dispatcher, 0),
		handlers:    make([]Handle, 0),
		doneCh:      make(chan struct{}),
	}
	go instance.timerChecker()
	go instance.flush()
	defaultLogger = instance
	return defaultLogger
}

func Fatal(ctx context.Context, tag string, err error, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Fatal(ctx, tag, err, fields...)
}

func Error(ctx context.Context, tag string, err error, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Error(ctx, tag, err, fields...)
}

func Warning(ctx context.Context, tag string, err error, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Warning(ctx, tag, err, fields...)
}

func Info(ctx context.Context, tag string, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Info(ctx, tag, fields...)
}

func Notice(ctx context.Context, tag string, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Notice(ctx, tag, fields...)
}

func Debug(ctx context.Context, tag string, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Debug(ctx, tag, fields...)
}

func Trace(ctx context.Context, tag string, fields ...entry.Field) {
	initDefaultLogger()
	defaultLogger.Trace(ctx, tag, fields...)
}

func AddHandle(ctx context.Context, handle Handle) {
	initDefaultLogger()
	defaultLogger.AddHandle(ctx, handle)
}

type logger struct {
	//ctx when ctx.Done(),flush logs
	ctx context.Context

	//cache log Entry
	logsCh chan entry.Entry

	//dispatchers for Dispatcher Dispatch  log Entry
	dispatchers []dispatcher.Dispatcher

	//level only log < Level.Code()'s Entry
	level level.Level

	//flush every 1 Seconds or len(logsCh)/cap(logsCh) > 0.8
	flushCh chan struct{}

	//handlers append field to entry
	handlers []Handle

	//doneCh notify
	doneCh chan struct{}
}

func (l *logger) Stop() {
	//不在接受日志
	close(l.logsCh)
	//刷新日志
	l.Flush()
	//通知关闭
	close(l.doneCh)
	//关闭刷新通道
	close(l.flushCh)
}

func (l *logger) AddHandle(ctx context.Context, handle Handle) {
	l.handlers = append(l.handlers, handle)
}

func (l *logger) AddDispatcher(dispatchers ...dispatcher.Dispatcher) {
	l.dispatchers = append(l.dispatchers, dispatchers...)
}

func (l *logger) SetLevel(level level.Level) {
	l.level = level
}

func (l *logger) Fatal(ctx context.Context, tag string, err error, fields ...entry.Field) {
	l.logger(ctx, level.LevelFatal, tag, err, fields...)
}

func (l *logger) Error(ctx context.Context, tag string, err error, fields ...entry.Field) {
	l.logger(ctx, level.LevelError, tag, err, fields...)
}

func (l *logger) Warning(ctx context.Context, tag string, err error, fields ...entry.Field) {
	l.logger(ctx, level.LevelWarning, tag, err, fields...)
}

func (l *logger) Info(ctx context.Context, tag string, fields ...entry.Field) {
	l.logger(ctx, level.LevelInfo, tag, nil, fields...)
}

func (l *logger) Notice(ctx context.Context, tag string, fields ...entry.Field) {
	l.logger(ctx, level.LevelNotice, tag, nil, fields...)
}

func (l *logger) Debug(ctx context.Context, tag string, fields ...entry.Field) {
	l.logger(ctx, level.LevelDebug, tag, nil, fields...)
}

func (l *logger) Trace(ctx context.Context, tag string, fields ...entry.Field) {
	l.logger(ctx, level.LevelTrace, tag, nil, fields...)
}

func (l *logger) logger(ctx context.Context, curLevel level.Level, tag string, err error, fields ...entry.Field) {
	if curLevel.Code() > l.level.Code() {
		return
	}
	e := entry.NewMapEntry(len(fields)+10, tag, curLevel)

	if curLevel.Code() <= level.LevelWarning.Code() {
		e.WithField("stack", GetStack(1, 16)) //string
	}

	e.WithField("errorMsg", "")
	for _, handle := range l.handlers {
		e.WithFields(handle(ctx, e)...)
	}
	e.WithFields(fields...)

	if err != nil {
		e.WithField("errorMsg", err.Error())
	}

	//@no block
	select {
	case l.logsCh <- e:
		l.check()
	default:
		//chan overflow
		l.Flush()
		go func() {
			l.logsCh <- e
		}()
	}
}

func (l *logger) Flush() {
	//no block
	select {
	case l.flushCh <- struct{}{}:
	default:
	}
}

func (l *logger) flush() {
	returnFlag := false
	for {
		if returnFlag {
			return
		}
		select {
		case _, ok := <-l.flushCh:
			if !ok {
				returnFlag = true
			}
			entries := make([]entry.Entry, 0, len(l.logsCh))
			flag := true
			for flag && len(entries) < cap(entries) {
				select {
				case e := <-l.logsCh:
					entries = append(entries, e)
				default:
					//chan空了
					flag = false
				}
			}
			if len(entries) == 0 {
				continue
			}
			for _, instance := range l.dispatchers {
				instance.Dispatch(entries)
			}
		case <-l.doneCh:
			return
		case <-l.ctx.Done():
			return
		}
	}
}

func (l *logger) timerChecker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.Flush()
		case <-l.ctx.Done():
			l.Flush() //@todo 安全退出
		}
	}
}

func (l *logger) check() {
	if len(l.logsCh)/cap(l.logsCh)*100 > 75 {
		l.Flush()
	}
}

func GetStack(skip int, depth int) string {
	buf := bytes.NewBufferString("")
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs)
	for _, pc := range pcs[:n] {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		//logger自己的stack不记录
		if strings.Contains(file, "neptune/logger") {
			continue
		}
		buf.WriteString(fmt.Sprintf("%s %d\n", file, line))
	}
	return buf.String()
}

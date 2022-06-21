package dispatcher

import (
	"bytes"
	"io"

	"github.com/no-mole/neptune/logger/formatter"

	"github.com/no-mole/neptune/logger/tagger"

	"github.com/no-mole/neptune/logger/entry"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	Registry("file", func(formatter formatter.Formatter, tagger tagger.Tagger, conf *Config) Dispatcher {
		writerConfig := &FileWriterConfig{
			Filename:   conf.GetString("fileName", "app.log"),
			MaxSize:    conf.GetInt("maxSize", 100),
			MaxBackups: conf.GetInt("maxBackups", 10),
			MaxAge:     conf.GetInt("maxAge", 15),
		}
		return NewFileDispatcher(writerConfig, formatter, tagger)
	})
}

func NewFileDispatcher(config *FileWriterConfig, formatter formatter.Formatter, tagger tagger.Tagger) Dispatcher {
	return &FileDispatcher{
		writer: fileWriter(config),
		Helper: NewHelper(formatter, tagger),
	}
}

type FileDispatcherConfig struct {
	Filename   string
	MaxSize    int  //最大M数，超过则切割
	MaxBackups int  //最大文件保留数，超过就删除最老的日志文件
	MaxAge     int  //保存30天
	Compress   bool //是否压缩
}

type FileDispatcher struct {
	Helper
	writer io.Writer
}

func (f *FileDispatcher) Dispatch(entries []entry.Entry) {
	buf := bytes.NewBuffer([]byte{})
	for _, e := range entries {
		if !f.Match(e.GetTag()) {
			continue
		}
		buf.Write(f.Format(e))
		buf.Write(lineBreak)
	}
	_, _ = f.writer.Write(buf.Bytes())
}

var _ Dispatcher = &FileDispatcher{}

type FileWriterConfig struct {
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
}

func fileWriter(config *FileWriterConfig) io.Writer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,    //最大M数，超过则切割
		MaxBackups: config.MaxBackups, //最大文件保留数，超过就删除最老的日志文件
		MaxAge:     config.MaxAge,     //保存30天
		Compress:   false,             //是否压缩
	}
	return lumberJackLogger
}

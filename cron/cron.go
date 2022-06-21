package cron

import (
	"context"
	"fmt"

	"github.com/no-mole/neptune/logger"
	"github.com/robfig/cron/v3"
)

type Manger struct {
	cron         *cron.Cron
	add          chan Job
	remove       chan string
	Closed       chan struct{}
	entryMapping map[string]cron.EntryID
	errChan      chan error
}

type Option func(m *Manger)

func WithErrChan(ch chan error) Option {
	return func(m *Manger) {
		m.errChan = ch
	}
}

type Job interface {
	Key() string
	Spec() string
	cron.Job
}

type cronLogger struct{}

func (l cronLogger) Info(msg string, keysAndValues ...interface{}) {
	logger.Info(context.Background(), "crontab", logger.WithField("msg", fmt.Sprintf(msg, keysAndValues...)))
}

func (l cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	logger.Error(context.Background(), "crontab", err, logger.WithField("msg", fmt.Sprintf(msg, keysAndValues...)))
}

func New(opts ...Option) *Manger {
	m := &Manger{
		cron:         cron.New(cron.WithLogger(cronLogger{})),
		add:          make(chan Job),
		remove:       make(chan string),
		Closed:       make(chan struct{}, 1),
		entryMapping: map[string]cron.EntryID{},
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manger) Start() {
	m.cron.Start()
	go func() {
		for {
			select {
			case job := <-m.add:
				entryId, err := m.cron.AddJob(job.Spec(), job)
				if err != nil {
					if m.errChan != nil {
						m.errChan <- err
					}
					continue
				}
				if oldEntryId, ok := m.entryMapping[job.Key()]; ok {
					m.cron.Remove(oldEntryId)
				}
				m.entryMapping[job.Key()] = entryId
			case key := <-m.remove:
				if entryId, ok := m.entryMapping[key]; ok {
					m.cron.Remove(entryId)
					delete(m.entryMapping, key)
				}
			case <-m.Closed:
				m.cron.Stop()
				m.entryMapping = map[string]cron.EntryID{}
				close(m.add)
				close(m.remove)
				return
			}
		}
	}()
}

func (m *Manger) Add(job Job) {
	m.add <- job
}

func (m *Manger) Remove(key string) {
	m.remove <- key
}

func (m *Manger) Stop() {
	close(m.Closed)
}

func (m *Manger) Entries() []cron.Entry {
	return m.cron.Entries()
}

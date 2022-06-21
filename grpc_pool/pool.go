package grpc_pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/no-mole/neptune/logger"

	"google.golang.org/grpc"

	"google.golang.org/grpc/connectivity"
)

const (
	defaultCycleMonitorTicker = 5 * time.Second
	defaultChannelCap         = 20
)

var (
	ClosedErr = errors.New("grpc pool has closed")
)

type Pool interface {
	Get() (Conn, error)
	Restore(Conn)
	Close()
}

type pool struct {
	sync.WaitGroup
	ctx                context.Context
	cancel             context.CancelFunc
	ticker             *time.Ticker
	builder            Builder
	pool               *stack
	storage            map[string]*rpcConn
	toRecycle          map[string]time.Time
	reStorage          chan *rpcConn
	buffer             chan *rpcConn
	activity           chan byte
	closed             chan bool
	currentConn        int           //连接池里当前连接数
	maxStreamsPerConn  int           //限制每个单连接的并发流数量
	maxIdle            int           //最小空闲
	maxActive          int           //最大活跃
	maxConnIdleSeconds time.Duration //最大连接空闲时间
	maxWaitConnTime    time.Duration //最大等待连接时间
}

// Option optional configs
type Option func(*Options)

type Options struct {
	// 最大空闲连接数.
	maxIdle int

	// 最大活跃连接数。
	maxActive int

	// 限制每个单连接的并发流数量
	maxStreamsPerConn int

	// 最大连接空闲时间
	maxConnIdleSeconds time.Duration

	// 最大等待连接时间
	maxWaitConnTime time.Duration

	dialOptions []grpc.DialOption
}

func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(o *Options) {
		o.dialOptions = append(o.dialOptions, opts...)
	}
}

func WithMaxIdle(n int) Option {
	return func(o *Options) {
		o.maxIdle = n
	}
}

func WithMaxActive(n int) Option {
	return func(o *Options) {
		o.maxActive = n
	}
}

func WithMaxConcurrentStreams(n int) Option {
	return func(o *Options) {
		o.maxStreamsPerConn = n
	}
}

func WithConnIdleSeconds(seconds time.Duration) Option {
	return func(o *Options) {
		o.maxConnIdleSeconds = seconds
	}
}

func WithWaitConn(millisecond time.Duration) Option {
	return func(o *Options) {
		o.maxWaitConnTime = millisecond
	}
}

func newPool(builder Builder, opt *Options) (Pool, error) {
	if builder == nil {
		return nil, errors.New("builder is null")
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &pool{
		ctx:                ctx,
		cancel:             cancel,
		ticker:             time.NewTicker(defaultCycleMonitorTicker),
		builder:            builder,
		storage:            make(map[string]*rpcConn),
		toRecycle:          make(map[string]time.Time),
		reStorage:          make(chan *rpcConn, defaultChannelCap),
		buffer:             make(chan *rpcConn, defaultChannelCap),
		activity:           make(chan byte, defaultChannelCap),
		closed:             make(chan bool),
		maxStreamsPerConn:  opt.maxStreamsPerConn,
		maxActive:          opt.maxActive,
		maxIdle:            opt.maxIdle,
		maxConnIdleSeconds: opt.maxConnIdleSeconds,
		maxWaitConnTime:    opt.maxWaitConnTime,
	}

	stack := &stack{
		values: make([]*rpcConn, 0),
	}
	for i := 0; i < opt.maxIdle; i++ {
		conn := newConn(builder, pool)
		stack.Push(conn)
	}
	pool.pool = stack

	go pool.Hold()
	return pool, nil
}

func (p *pool) Get() (Conn, error) {
	p.activity <- 1
	for {
		select {
		case <-p.ctx.Done():
			return nil, ClosedErr
		case rpcConn := <-p.buffer:
			if rpcConn.conn.GetState() != connectivity.Shutdown {
				p.Add(1)
				return rpcConn, nil
			} else {
				p.activity <- 1
			}
		}
	}
}

func (p *pool) Restore(c Conn) {
	if c == nil {
		return
	}
	p.reStorage <- c.(*rpcConn)
	p.Done()
}

func (p *pool) Close() {
	p.cancel()
	p.ticker.Stop()

	p.Wait()
	close(p.closed)

	//关闭正在回收的conn
	for id := range p.toRecycle {
		p.pool.Remove(id).conn.Close()
	}
	for !p.pool.Empty() {
		p.pool.Pop().conn.Close()
	}
}

func (p *pool) Hold() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(context.Background(), "err", fmt.Errorf("grpc pool hold %v", r))
			p.Hold()
		}
	}()

	for {
		select {
		case <-p.closed:
			return
		case <-p.ticker.C:
			for id, ts := range p.toRecycle {
				//超过conn的最大连接时间
				if time.Since(ts).Seconds() > p.maxConnIdleSeconds.Seconds() {
					if p.currentConn > p.maxIdle {
						conn := p.pool.Remove(id)
						conn.conn.Close()
						p.currentConn--
						delete(p.toRecycle, id)
					}
				}
			}
		case conn := <-p.reStorage: //放回池中
			//stream 满了的conn
			if c, ok := p.storage[conn.id]; ok {
				//释放一个conn
				delete(p.storage, c.id)
				c.streams--
				p.pool.Push(c)
				continue
			}
			if conn.streams--; conn.streams == 0 {
				if p.currentConn > p.maxIdle {
					p.toRecycle[conn.id] = time.Now()
				}
			}

		case <-p.activity:
			retry := 5
		GET:
			conn := p.pool.Peek()
			if conn == nil {
				//没超过最大活跃数
				if p.currentConn <= p.maxActive {
					newConn := newConn(p.builder, p)
					p.pool.Push(newConn)
					p.currentConn++
					goto GET
				}
				//等待其他连接释放
				for retry > 0 {
					retry--
					<-time.After(p.maxWaitConnTime)
					goto GET
				}
				continue
			}
			if conn.conn.GetState() == connectivity.Shutdown {
				// 检查关掉shutdown的conn不会panic
				conn = p.pool.Pop()
				conn.Close()
				p.currentConn--
				goto GET
			}

			if _, ok := p.toRecycle[conn.id]; ok {
				delete(p.toRecycle, conn.id)
			}
			//小于最大stream
			if conn.streams+1 <= p.maxStreamsPerConn {
				goto PUT
			}
			//超出最大stream，从pool中取出
			p.storage[conn.id] = p.pool.Pop()
			goto GET
		PUT:
			conn.streams++
			p.buffer <- conn
		}
	}
}

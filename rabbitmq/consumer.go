package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

const consumerReconnectDelay = 3 * time.Second

type Consumer func(msg amqp.Delivery) error

type RabbitMQConsumer struct {
	ctx                context.Context
	workers            int64
	eg                 *errgroup.Group
	ErrChan            chan error
	clientName         string
	queueDeclareParams *QueueDeclareParams
	queueBindParams    []*QueueBindParams
	consumeParams      *ConsumeParams
}

type QueueDeclareParams struct {
	Queue      string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type QueueBindParams struct {
	QueueName    string
	RoutingKey   string
	ExchangeName string
	NoWait       bool
	Args         amqp.Table
}

type ConsumeParams struct {
	QueueName    string
	ConsumerName string
	AutoAck      bool
	Exclusive    bool
	NoLocal      bool
	NoWait       bool
	Args         amqp.Table
}

func NewRabbitMQConsumer(workSize int64, rabbitModel string,
	queueDeclareParams *QueueDeclareParams, consumeParams *ConsumeParams, queueBindParams ...*QueueBindParams) *RabbitMQConsumer {
	g, ctx := errgroup.WithContext(context.Background())
	r := &RabbitMQConsumer{
		workers:            workSize,
		eg:                 g,
		ctx:                ctx,
		clientName:         rabbitModel,
		queueDeclareParams: queueDeclareParams,
		queueBindParams:    queueBindParams,
		consumeParams:      consumeParams,
	}
	r.ErrChan = make(chan error, 10)
	return r
}

func (f *RabbitMQConsumer) Working(consumer Consumer) (err error) {
	_, exist := Client.GetClient(f.clientName)
	if !exist {
		return errors.New("rabbitmq client is not exist，client name is " + f.clientName)
	}

	for i := 0; i < int(f.workers); i++ {
		f.eg.Go(func() (err error) {
			for {
				if err = f.consume(consumer); err != nil {
					return err
				}
				if err = f.waitReconnect(); err != nil {
					return err
				}
			}
		})
	}
	err = f.eg.Wait()
	if err != nil {
		f.ErrChan <- err
	}
	return nil
}

func (f *RabbitMQConsumer) consume(consumer Consumer) error {
	cli, exist := Client.GetClient(f.clientName)
	if !exist {
		return errors.New("rabbitmq client is not exist，client name is " + f.clientName)
	}
	ch, err := cli.Channel()
	if err != nil {
		return nil
	}
	defer ch.Close()

	if err = f.prepareChannel(ch); err != nil {
		return nil
	}
	msgCh, err := ch.Consume(f.consumeParams.QueueName, f.consumeParams.ConsumerName, f.consumeParams.AutoAck,
		f.consumeParams.Exclusive, f.consumeParams.NoLocal, f.consumeParams.NoWait, f.consumeParams.Args)
	if err != nil {
		return nil
	}

	for {
		select {
		case <-f.ctx.Done():
			return f.ctx.Err()
		case msg, ok := <-msgCh:
			if !ok {
				return nil
			}
			if err := consumer(msg); err != nil {
				return err
			}
			_ = msg.Ack(false)
		}
	}
}

func (f *RabbitMQConsumer) prepareChannel(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(f.queueDeclareParams.Queue, f.queueDeclareParams.Durable, f.queueDeclareParams.AutoDelete,
		f.queueDeclareParams.Exclusive, f.queueDeclareParams.NoWait, f.queueDeclareParams.Args)
	if err != nil {
		return err
	}

	for _, v := range f.queueBindParams {
		err = ch.QueueBind(v.QueueName, v.RoutingKey, v.ExchangeName, v.NoWait, v.Args)
		if err != nil {
			return err
		}
	}

	return ch.Qos(8, 0, false)
}

func (f *RabbitMQConsumer) waitReconnect() error {
	timer := time.NewTimer(consumerReconnectDelay)
	defer timer.Stop()

	select {
	case <-f.ctx.Done():
		return f.ctx.Err()
	case <-timer.C:
		return nil
	}
}

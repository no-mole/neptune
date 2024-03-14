package rabbitmq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

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

	cli, exist := Client.GetClient(f.clientName)
	if !exist {
		return errors.New("rabbitmq client is not exist，client name is " + f.clientName)
	}
	ch, err := cli.Channel()
	if err != nil {
		return
	}

	_, err = ch.QueueDeclare(f.queueDeclareParams.Queue, f.queueDeclareParams.Durable, f.queueDeclareParams.AutoDelete,
		f.queueDeclareParams.Exclusive, f.queueDeclareParams.NoWait, f.queueDeclareParams.Args)
	if err != nil {
		return
	}

	for _, v := range f.queueBindParams {
		err = ch.QueueBind(v.QueueName, v.RoutingKey, v.ExchangeName,
			f.queueDeclareParams.NoWait, f.queueDeclareParams.Args)
		if err != nil {
			return err
		}
	}

	_ = ch.Close()

	for i := 0; i < int(f.workers); i++ {
		f.eg.Go(func() (err error) {
			ch, err := cli.Channel()
			if err != nil {
				return
			}
			defer ch.Close()

			err = ch.Qos(8, 0, false)
			if err != nil {
				return
			}
			msgCh, err := ch.Consume(f.consumeParams.QueueName, f.consumeParams.ConsumerName, f.consumeParams.AutoAck,
				f.consumeParams.Exclusive, f.consumeParams.NoLocal, f.consumeParams.NoWait, f.queueDeclareParams.Args)
			if err != nil {
				return
			}
			for {
				select {
				case <-f.ctx.Done():
					return f.ctx.Err()
				case msg := <-msgCh:
					err := consumer(msg)
					if err != nil {
						return err
					}
					_ = msg.Ack(false)
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

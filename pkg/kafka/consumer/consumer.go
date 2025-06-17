package consumer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/avast/retry-go"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"google.golang.org/protobuf/proto"

	k "github.com/COTBU/sotbi.lib/pkg/kafka"
	"github.com/COTBU/sotbi.lib/pkg/log"
)

const (
	MaxAttempts  = 5
	DelayTimeout = 200 * time.Millisecond
)

type ConsumerOptions struct {
	Brokers []string
	GroupID string

	Username string
	Password string

	// Cert опциональный параметр, в котором можно передать
	// тело сертификата, который будет использован при создании защищенного соединения
	// Если задан, значение параметра TLS игнорируется
	Cert string `exhaustruct:"optional"`
	// TLS опциональный параметр, если true,
	// то будет установлено защищенное соединение.
	// Если Сert при этом не указан, будут использованы
	// системные корневые сертификаты
	TLS bool `exhaustruct:"optional"`

	ReadEarliest bool
}

type consumer[T proto.Message] struct {
	reader      *kafka.Reader
	logger      log.Logger
	newInstance func() T
	handleFunc  func(context.Context, T) error
}

func NewConsumer[T proto.Message](
	newInstance func() T,
	handleFunc func(context.Context, T) error,
	opts *ConsumerOptions,
	optFunc ...consumerOptionFunc,
) (*consumer[T], error) {
	customOpts := &kafkaFuncOpts{
		minBytes:       1e3,
		maxBytes:       10e6,
		fetchMaxWait:   10 * time.Second,
		commitInterval: 0,
		dialerTimeout:  time.Second,
	}

	for _, opt := range optFunc {
		opt(customOpts)
	}

	dialer := &kafka.Dialer{
		Timeout:   customOpts.dialerTimeout,
		DualStack: true,
	}

	dialer.TLS = k.GetTLSConfig(opts.TLS, opts.Cert)

	if opts.Username != "" && opts.Password != "" {
		mechanism, err := scram.Mechanism(
			scram.SHA512,
			opts.Username,
			opts.Password)
		if err != nil {
			return nil, err
		}

		dialer.SASLMechanism = mechanism
	}

	startOffset := kafka.LastOffset
	if opts.ReadEarliest {
		startOffset = kafka.FirstOffset
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:               opts.Brokers,
		Topic:                 customOpts.topic,
		GroupID:               opts.GroupID,
		WatchPartitionChanges: true,
		StartOffset:           startOffset,
		ErrorLogger:           customOpts.logger,
		Dialer:                dialer,
		MaxBytes:              customOpts.maxBytes,
		MinBytes:              customOpts.minBytes,
		MaxWait:               customOpts.fetchMaxWait,
		CommitInterval:        customOpts.commitInterval,
		ReadBackoffMin:        100 * time.Millisecond,
		ReadBackoffMax:        1 * time.Second,
		MaxAttempts:           MaxAttempts,
	}

	r := kafka.NewReader(readerConfig)

	customOpts.logger.Info(
		"Consumer created",
		"brokers",
		opts.Brokers,
		"groupID",
		opts.GroupID,
		"topic",
		customOpts.topic,
		"fetchMaxWait",
		customOpts.fetchMaxWait,
	)

	return &consumer[T]{
		reader:      r,
		logger:      customOpts.logger,
		newInstance: newInstance,
		handleFunc:  handleFunc,
	}, nil
}

func (c *consumer[T]) fetchMessage(ctx context.Context) (*k.Message[T], error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, err

		case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
			return nil, err

		default:
			return nil, fmt.Errorf("%w | %w", k.ErrFetchMessage, err)
		}
	}

	resMsg := &k.Message[T]{
		Key:      msg.Key,
		Value:    c.newInstance(),
		Headers:  k.KafkaHeadersToHeaders(msg.Headers),
		RawValue: msg.Value,
		Msg:      &msg,
	}

	var value proto.Message = c.newInstance()

	if len(msg.Value) == 0 {
		return resMsg, fmt.Errorf("%w | %w", k.ErrValueUnmarshalling, k.ErrEmptyValue)
	}

	data := msg.Value

	if data[0] == 0x0 {
		data = data[5:]
		msgIndexBytesLen := len(k.ToMessageIndexBytes(value.ProtoReflect().Descriptor()))
		data = data[msgIndexBytesLen:]
	}

	if err := proto.Unmarshal(data, value); err != nil {
		return resMsg, fmt.Errorf("%w | %w", k.ErrValueUnmarshalling, err)
	}

	var ok bool

	resMsg.Value, ok = value.(T)
	if !ok {
		return nil, fmt.Errorf("value is not of type %T", value)
	}

	return resMsg, nil
}

func (c *consumer[T]) commitMessage(ctx context.Context, msg ...*k.Message[T]) error {
	if len(msg) < 1 {
		return nil
	}

	cMsg := make([]kafka.Message, 0, len(msg))

	for _, message := range msg {
		cMsg = append(cMsg, *message.Msg)
	}

	if err := c.reader.CommitMessages(ctx, cMsg...); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return err

		case errors.Is(err, io.EOF),
			errors.Is(err, io.ErrUnexpectedEOF),
			errors.Is(err, io.ErrClosedPipe):
			return err

		default:
			return fmt.Errorf("%w | %w", k.ErrCommitMessage, err)
		}
	}

	return nil
}

func (c *consumer[T]) Close() error {
	err := c.reader.Close()
	if err != nil {
		return fmt.Errorf("%w | %w", k.ErrCloseConsumer, err)
	}

	return nil
}

func (c *consumer[T]) handleMessage(ctx context.Context, msg T) error {
	return c.handleFunc(ctx, msg)
}

func (c *consumer[T]) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var msg *k.Message[T]

			var err error

			err = retry.Do(
				func() error {
					msg, err = c.fetchMessage(ctx)
					if err != nil {
						return fmt.Errorf("consumer.FetchMessage: %w", err)
					}

					return nil
				},
				retry.Attempts(MaxAttempts),
				retry.RetryIf(func(err error) bool {
					return errors.Is(err, k.ErrFetchMessage)
				}),
			)
			if err != nil {
				c.logger.Error("failed to fetch message", err)

				continue
			}

			err = retry.Do(
				func() error {
					if err := c.handleMessage(ctx, msg.Value); err != nil {
						return fmt.Errorf("c.consumer.HandleMessage: %w", err)
					}

					return nil
				},
				retry.Attempts(MaxAttempts),
				retry.Delay(DelayTimeout),
				retry.OnRetry(func(n uint, err error) {
					c.logger.Error(
						"failed to handle message",
						msg,
						fmt.Errorf("attempt %d: %w", n, err),
					)
				}))
			if err != nil {
				c.logger.Error("failed to handle message", err)

				continue
			}

			err = c.commitMessage(ctx, msg)
			if err != nil {
				c.logger.Error("failed to commit message", err)
			}
		}
	}
}

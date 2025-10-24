package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"google.golang.org/protobuf/proto"

	k "github.com/SOTBI-LLC/sotbi.lib/pkg/kafka"
	"github.com/SOTBI-LLC/sotbi.lib/pkg/log"
	"github.com/SOTBI-LLC/sotbi.lib/pkg/log/slog"
)

type Transport struct {
	tr *kafka.Transport
}

type TransportOptions struct {
	DialerTimeout time.Duration
	Username      string
	Password      string
	// Cert опциональный параметр, в котором можно передать
	// тело сертификата, который будет использован при создании защищенного соединения
	// Если задан, значение параметра TLS игнорируется
	Cert string `exhaustruct:"optional"`
	// TLS опциональный параметр, если true,
	// то будет установлено защищенное соединение.
	// Если Сert при этом не указан, будут использованы
	// системные корневые сертификаты
	TLS bool `exhaustruct:"optional"`
}

func NewTransport(opts *TransportOptions) (*Transport, error) {
	transport := &kafka.Transport{
		DialTimeout: opts.DialerTimeout,
		TLS:         k.GetTLSConfig(opts.TLS, opts.Cert),
		SASL:        nil,
	}

	if opts.Password != "" && opts.Username != "" {
		mechanism, err := scram.Mechanism(scram.SHA512, opts.Username, opts.Password)
		if err != nil {
			return nil, err
		}

		transport.SASL = mechanism
	}

	return &Transport{
		tr: transport,
	}, nil
}

type producer[T proto.Message] struct {
	writer         *kafka.Writer
	schemaRegistry *SchemaRegistry
	cluster        string
}

type ProducerOptions struct {
	Brokers         []string
	Logger          log.Logger
	Transport       *Transport // shared resource
	requiredAcks    RequiredAcks
	balancer        kafka.Balancer
	batchSize       int
	batchTimeout    time.Duration
	schemaRegistry  *SchemaRegistry
	maxAttempts     int
	writeBackoffMin time.Duration
	writeBackoffMax time.Duration
	batchBytes      int64
	compression     Compression
}

func newProducer[T proto.Message](
	cluster string,
	opts *ProducerOptions,
	optFunc ...func(opts *ProducerOptions),
) *producer[T] {
	opts.batchSize = defaultBatchSize
	opts.batchTimeout = defaultBatchTimeout
	opts.maxAttempts = defaultMaxAttempts
	opts.writeBackoffMin = defaultWriteBackoffMin
	opts.writeBackoffMax = defaultWriteBackoffMax
	opts.compression = defaultCompression

	for _, opt := range optFunc {
		opt(opts)
	}

	if opts.schemaRegistry != nil {
		if err := opts.schemaRegistry.registerSchemas(); err != nil && opts.Logger != nil {
			opts.Logger.Error("error while register schema registry", err)
		}
	}

	if len(opts.Brokers) < 1 {
		panic("opts.Brokers must not be empty")
	}

	if opts.Logger == nil {
		opts.Logger = slog.New("error")
	}

	var tr *kafka.Transport
	if opts.Transport != nil {
		tr = opts.Transport.tr
	}

	return &producer[T]{
		writer: &kafka.Writer{
			BatchSize:       opts.batchSize,
			BatchTimeout:    opts.batchTimeout,
			Addr:            kafka.TCP(opts.Brokers...),
			ErrorLogger:     opts.Logger,
			Transport:       tr,
			RequiredAcks:    kafka.RequiredAcks(opts.requiredAcks),
			Balancer:        opts.balancer,
			MaxAttempts:     opts.maxAttempts,
			WriteBackoffMin: opts.writeBackoffMin,
			WriteBackoffMax: opts.writeBackoffMax,
			BatchBytes:      opts.batchBytes,
			Compression:     kafka.Compression(opts.compression),
		},
		schemaRegistry: opts.schemaRegistry,
		cluster:        cluster,
	}
}

func NewProducer[T proto.Message](
	opts *ProducerOptions,
	optFunc ...func(opts *ProducerOptions),
) *producer[T] { //nolint:revive
	return newProducer[T]("default", opts, optFunc...)
}

func (p *producer[T]) Produce(
	ctx context.Context,
	msg ...*k.Message[T],
) error {
	if len(msg) < 1 {
		return nil
	}

	kafkaMessages := make([]kafka.Message, 0, len(msg))

	for _, message := range msg {
		protoData, err := proto.Marshal(message.Value)
		if err != nil {
			return fmt.Errorf("%w | %w", k.ErrMarshalValue, err)
		}

		if p.schemaRegistry != nil {
			schemaID := p.schemaRegistry.getSchemaID(message.Topic)
			if schemaID != nil {
				msgIndexBytes := k.ToMessageIndexBytes(message.Value.ProtoReflect().Descriptor())

				protoData, err = addSchemaIDPrefix(*schemaID, append(msgIndexBytes, protoData...))
				if err != nil {
					return fmt.Errorf("%w | %w", k.ErrMarshalValue, err)
				}
			}
		}

		kafkaHeaders := make([]kafka.Header, 0, len(message.Headers))

		for _, header := range message.Headers {
			kafkaHeaders = append(kafkaHeaders, kafka.Header{
				Key:   header.Key,
				Value: header.Value,
			})
		}

		topic := message.Topic
		if topic == "" {
			return fmt.Errorf("topic must be specified")
		}

		kafkaMessages = append(kafkaMessages, kafka.Message{
			Topic:   topic,
			Key:     message.Key,
			Headers: kafkaHeaders,
			Value:   protoData,
		})
	}

	err := p.writer.WriteMessages(ctx, kafkaMessages...)
	if err != nil {
		err = fmt.Errorf("%w | %w", k.ErrWriteMessage, err)
	}

	return err
}

func (p *producer[T]) Close() error {
	err := p.writer.Close()
	if err != nil {
		return fmt.Errorf("%w | %w", k.ErrCloseProducer, err)
	}

	return nil
}

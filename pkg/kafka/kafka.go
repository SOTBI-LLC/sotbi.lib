package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	ErrCloseConsumer      = errors.New("unable to close consumer")
	ErrValueUnmarshalling = errors.New("unable to unmarshal message value")
	ErrFetchMessage       = errors.New("unable to fetch message")
	ErrCommitMessage      = errors.New("unable to commit message")
	ErrEmptyValue         = errors.New("empty message value")
	ErrCloseProducer      = errors.New("unable to close producer")
	ErrMarshalValue       = errors.New("unable to marshal message value")
	ErrWriteMessage       = errors.New("unable to write message")
)

type Header struct {
	Key   string
	Value []byte
}

type Message[T proto.Message] struct {
	// Topic overrides the name of the topic specified when generating the producer (useful for testing)
	Topic    string
	Key      []byte
	Value    T
	Headers  []Header
	RawValue []byte
	Msg      *kafka.Message
}

func (m *Message[T]) GetKey() []byte {
	return m.Key
}

type Producer[T proto.Message] interface {
	Produce(context.Context, ...*Message[T]) error
	Close() error
}

func GetTLSConfig(useTLS bool, cert string) *tls.Config {
	if !useTLS && cert == "" {
		return nil
	}

	if useTLS && cert == "" {
		return &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
			RootCAs:            nil,
		}
	}

	// whatever && cert != ""
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM([]byte(cert))

	return &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		RootCAs:            certs,
	}
}

func KafkaHeadersToHeaders(headers []kafka.Header) []Header {
	if headers == nil {
		return nil
	}

	res := make([]Header, 0, len(headers))

	for _, h := range headers {
		res = append(res, Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	return res
}

func toMessageIndexes(descriptor protoreflect.Descriptor, count int) []int {
	index := descriptor.Index()

	switch v := descriptor.Parent().(type) {
	case protoreflect.FileDescriptor:
		msgIndexes := make([]int, count+1)
		msgIndexes[0] = index

		return msgIndexes[0:1]
	default:
		msgIndexes := toMessageIndexes(v, count+1)

		return append(msgIndexes, index)
	}
}

func ToMessageIndexBytes(descriptor protoreflect.Descriptor) []byte {
	if descriptor.Index() == 0 {
		if _, ok := descriptor.Parent().(protoreflect.FileDescriptor); ok {
			return []byte{0}
		}
	}

	msgIndexes := toMessageIndexes(descriptor, 0)
	buf := make([]byte, (1+len(msgIndexes))*binary.MaxVarintLen64)
	length := binary.PutVarint(buf, int64(len(msgIndexes)))

	for _, element := range msgIndexes {
		length += binary.PutVarint(buf[length:], int64(element))
	}

	return buf[0:length]
}

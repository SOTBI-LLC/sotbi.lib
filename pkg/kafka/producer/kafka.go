//nolint:godot,lll
package producer

import (
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/log"
)

type RequiredAcks int

const (
	RequireNone RequiredAcks = 0
	RequireOne  RequiredAcks = 1
	RequireAll  RequiredAcks = -1
)

type Compression int8

const (
	None   Compression = 0
	Gzip   Compression = 1
	Snappy Compression = 2
	Lz4    Compression = 3
	Zstd   Compression = 4
)

const (
	defaultBatchTimeout    = 10 * time.Millisecond
	maxBatchTimeout        = 200 * time.Millisecond
	defaultBatchSize       = 1
	defaultMaxAttempts     = 3
	defaultWriteBackoffMin = 100 * time.Millisecond //nolint:revive
	defaultWriteBackoffMax = 1 * time.Second        //nolint:revive
	defaultCompression     = 0
)

// WithRequiredAcks
// RequireNone - означает, что сообщение считается успешно записанным в Kafka, если производитель сумел отправить его по сети.
// RequireOne - означает, что ведущая реплика в момент получения сообщения и записи его в файл данных раздела (но не обязательно на диск) отправила подтверждение или сообщение об ошибке.
// RequireAll - означает, что ведущая реплика, прежде чем отправлять подтверждение или сообщение об ошибке, дождется получения сообщения всеми согласованными репликами.
func WithRequiredAcks(requiredAcks RequiredAcks) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.requiredAcks = requiredAcks
	}
}

// WithCompression
// None - без сжатия.
// Gzip - сжатие с помощью алгоритма Gzip.
// Snappy - сжатие с помощью алгоритма Snappy.
// Lz4 - сжатие с помощью алгоритма LZ4.
// Zstd - сжатие с помощью алгоритма Zstandard.
func WithCompression(сompression Compression) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.compression = сompression
	}
}

// WithBatchSize определяет количество сообщений, отправляемых в одном пакете.
// Значение по-умолчанию: 1.
func WithBatchSize(batchSize int) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.batchSize = batchSize
	}
}

// WithBatchBytes ограничение максимального размера запроса в байтах перед отправкой в партишн.
func WithBatchBytes(batchBytes int64) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.batchBytes = batchBytes
	}
}

// WithBatchTimeout задает максимальное время ожидания размера пакета перед публикацией,
// чтобы избежать длительного ожидания в топиках с низкой производительностью
// Используйте с осторожностью, т.к. BatchTimeout заблокирует Write до того момента, пока не накопится
// BatchSize или не пройдет указанное время ожидания BatchTimeout
// Учитывайте, что продюсер работает в синхронном режиме.
// Значение по-умолчанию: 10мс
// Рекомендуемое максимальное значение: 200мс
func WithBatchTimeout(batchTimeout time.Duration) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.batchTimeout = batchTimeout
	}
}

// WithBalancer
// Интерфейс kafka.Balancer предоставляет абстракцию логики распределения сообщений,
// используемой экземплярами Writer для маршрутизации сообщений к доступным разделам на кластере Kafka.
func WithBalancer(balancer kafka.Balancer) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.balancer = balancer
	}
}

func WithMaxAttempts(maxAttempts int) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.maxAttempts = maxAttempts
	}
}

func WithWriteBackoffMin(backoffMin time.Duration) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.writeBackoffMin = backoffMin
	}
}

func WithLogger(log log.Logger) func(opts *ProducerOptions) {
	return func(opts *ProducerOptions) {
		opts.Logger = log
	}
}

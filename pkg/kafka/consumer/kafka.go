package consumer

import (
	"time"

	"github.com/COTBU/sotbi.lib/pkg/log"
)

type kafkaFuncOpts struct {
	// Минимальный объем данных, получаемых от брокера при извлечении записей.
	// Установка большого значения, может привести к задержке доставки,
	// т.к у брокера будет недостаточно данных для удовлетворения заданного минимума.
	minBytes int
	// Максимальный размер пакета, который может принять потребитель.
	// Брокер обрезает сообщение, чтобы удовлетворить этому максимуму,
	// поэтому выберите значение, достаточно высокое для вашего самого "тяжёлого" сообщения.
	maxBytes int
	// Максимальное время ожидания достаточного объёма данных (minBytes) при получении пакетов сообщений из kafka.
	fetchMaxWait   time.Duration
	commitInterval time.Duration
	dialerTimeout  time.Duration
	logger         log.Logger
	topic          string
}

type consumerOptionFunc func(opts *kafkaFuncOpts)

// Минимальный объем данных, получаемых от брокера при извлечении записей.
// Установка большого значения, может привести к задержке доставки,
// т.к у брокера будет недостаточно данных для удовлетворения заданного минимума.
func WithMinBytes(value int) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.minBytes = value
	}
}

// Максимальный размер пакета, который может принять потребитель.
// Брокер обрезает сообщение, чтобы удовлетворить этому максимуму,
// поэтому выберите значение, достаточно высокое для вашего самого "тяжёлого" сообщения.
func WithMaxBytes(value int) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.maxBytes = value
	}
}

// Максимальное время ожидания достаточного объёма данных (minBytes) при получении пакетов сообщений из kafka.
func WithFetchMaxWait(value time.Duration) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.fetchMaxWait = value
	}
}

// CommitInterval указывает интервал, через который commit фиксируются брокеру (CommitMessage)
// Если 0, фиксации будут обрабатываться синхронно (по умолчанию).
func WithCommitInterval(value time.Duration) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.commitInterval = value
	}
}

func WithLogger(log log.Logger) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.logger = log
	}
}

func WithTopic(topic string) consumerOptionFunc {
	return func(opts *kafkaFuncOpts) {
		opts.topic = topic
	}
}

package consumer

import (
	"context"
	"errors"
	"time"
)

// Operation — внутренний шаг consumer, результаты которого наблюдаются.
type Operation string

const (
	// OpFetch — получение сообщения из Kafka (сетевой вызов FetchMessage).
	OpFetch Operation = "fetch"
	// OpDecode — декодирование значения сообщения (без сетевых вызовов).
	OpDecode Operation = "decode"
	// OpHandler — вызов пользовательского обработчика сообщения.
	OpHandler Operation = "handler"
	// OpCommit — фиксация offset обработанного сообщения.
	OpCommit Operation = "commit"
)

// OperationResult — ограниченная классификация результата операции.
// Произвольные тексты ошибок в классификацию не попадают.
type OperationResult string

const (
	// ResultSuccess — операция завершилась без ошибки.
	ResultSuccess OperationResult = "success"
	// ResultError — операция завершилась ошибкой (не связанной с отменой).
	ResultError OperationResult = "error"
	// ResultCanceled — операция прервана отменой контекста или дедлайном.
	ResultCanceled OperationResult = "canceled"
)

// Observer получает результат каждой законченной попытки внутренней
// операции consumer (fetch/decode/handler/commit), включая неуспешные
// попытки внутренних ретраев, не выходя из Consume.
//
// Контракт:
//   - вызов синхронный: реализация обязана быть неблокирующей и быстрой;
//   - наблюдение не влияет на решение о retry, commit или обработке;
//   - длительность передаётся как time.Duration, payload сообщения не передаётся;
//   - классификация результата ограничена значениями OperationResult.
//
// Ограничение видимости: ретраи JoinGroup и координация группы внутри
// kafka-go не проходят через Observer — ошибки этой фазы видны только как
// ошибки fetch, когда FetchMessage возвращает ошибку. События fetch нельзя
// считать полным счётчиком событий JoinGroup.
type Observer interface {
	ObserveOperation(op Operation, result OperationResult, duration time.Duration)
}

// ObserverFunc — адаптер функции к Observer.
type ObserverFunc func(op Operation, result OperationResult, duration time.Duration)

func (f ObserverFunc) ObserveOperation(
	op Operation,
	result OperationResult,
	duration time.Duration,
) {
	f(op, result, duration)
}

// observeOperation публикует результат законченной попытки операции.
// При nil observer вызов сводится к одной проверке.
func (c *consumer[T]) observeOperation(op Operation, start time.Time, err error) {
	if c.observer == nil {
		return
	}

	result := ResultSuccess
	if err != nil {
		result = ResultError
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			result = ResultCanceled
		}
	}

	c.observer.ObserveOperation(op, result, time.Since(start))
}

package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	logmocks "github.com/SOTBI-LLC/sotbi.lib/pkg/log/mocks"

	pb "github.com/SOTBI-LLC/sotbi.lib/pkg/api/notification"
)

func validNotificationMessage(t *testing.T) kafka.Message {
	t.Helper()

	value, err := proto.Marshal(&pb.Notification{Subject: "observer-test"})
	require.NoError(t, err)

	return kafka.Message{Value: value}
}

// newTestConsumer собирает consumer с моком reader и observer без реального брокера.
func newTestConsumer(
	t *testing.T,
	reader messageReader,
	observer Observer,
	handleFunc func(context.Context, *pb.Notification) error,
) *consumer[*pb.Notification] {
	t.Helper()

	logger := logmocks.NewMockLogger(t)
	logger.EXPECT().Error(mock.Anything, mock.Anything).Maybe()
	logger.EXPECT().Warn(mock.Anything, mock.Anything).Maybe()
	logger.EXPECT().Printf(mock.Anything, mock.Anything).Maybe()

	return &consumer[*pb.Notification]{
		reader:      reader,
		logger:      logger,
		newInstance: func() *pb.Notification { return &pb.Notification{} },
		handleFunc:  handleFunc,
		observer:    observer,
	}
}

func noopHandle(_ context.Context, _ *pb.Notification) error {
	return nil
}

// Каждая внутренняя попытка fetch видна отдельным событием, включая canceled.
func TestObserver_FetchErrorAttemptsVisible(t *testing.T) {
	reader := newMockmessageReader(t)
	observer := NewMockObserver(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			calls++
			if calls <= MaxAttempts {
				return kafka.Message{}, errors.New("broker unavailable")
			}
			cancel()

			return kafka.Message{}, context.Canceled
		},
	).Times(MaxAttempts + 1)

	observer.EXPECT().ObserveOperation(OpFetch, ResultError, mock.Anything).Times(MaxAttempts)
	observer.EXPECT().ObserveOperation(OpFetch, ResultCanceled, mock.Anything).Once()
	// Ни decode, ни handler, ни commit при неполученном сообщении не выполняются:
	// незадекларированный вызов упадёт как unexpected method call.

	c := newTestConsumer(t, reader, observer, noopHandle)
	require.NoError(t, c.Consume(ctx))
}

// Ошибка декодирования наблюдается отдельно от успешного fetch и не ретраится.
func TestObserver_DecodeError(t *testing.T) {
	reader := newMockmessageReader(t)
	observer := NewMockObserver(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader.EXPECT().FetchMessage(mock.Anything).Return(
		kafka.Message{Value: []byte{0xff, 0x00, 0x01, 0x02}}, nil,
	).Once()
	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			cancel()

			return kafka.Message{}, context.Canceled
		},
	).Once()

	observer.EXPECT().ObserveOperation(OpFetch, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpDecode, ResultError, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpFetch, ResultCanceled, mock.Anything).Once()

	c := newTestConsumer(t, reader, observer, noopHandle)
	require.NoError(t, c.Consume(ctx))
}

// Каждая попытка handler видна отдельно; commit успешного сообщения наблюдается.
func TestObserver_HandlerAttemptsThenCommit(t *testing.T) {
	reader := newMockmessageReader(t)
	observer := NewMockObserver(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg := validNotificationMessage(t)

	reader.EXPECT().FetchMessage(mock.Anything).Return(msg, nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			cancel()

			return kafka.Message{}, context.Canceled
		},
	).Once()
	reader.EXPECT().CommitMessages(mock.Anything, []kafka.Message{msg}).Return(nil).Once()

	handlerCalls := 0
	handleFunc := func(_ context.Context, _ *pb.Notification) error {
		handlerCalls++
		if handlerCalls <= 2 {
			return errors.New("smtp failure")
		}

		return nil
	}

	observer.EXPECT().ObserveOperation(OpFetch, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpDecode, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpHandler, ResultError, mock.Anything).Times(2)
	observer.EXPECT().ObserveOperation(OpHandler, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpCommit, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpFetch, ResultCanceled, mock.Anything).Once()

	c := newTestConsumer(t, reader, observer, handleFunc)
	require.NoError(t, c.Consume(ctx))
}

// Commit со ошибкой и последующий успешный commit наблюдаются, не выходя из Consume.
func TestObserver_CommitErrorThenSuccess(t *testing.T) {
	reader := newMockmessageReader(t)
	observer := NewMockObserver(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first := validNotificationMessage(t)
	second := validNotificationMessage(t)

	reader.EXPECT().FetchMessage(mock.Anything).Return(first, nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).Return(second, nil).Once()
	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			cancel()

			return kafka.Message{}, context.Canceled
		},
	).Once()
	reader.EXPECT().
		CommitMessages(mock.Anything, []kafka.Message{first}).
		Return(errors.New("commit failed")).
		Once()
	reader.EXPECT().CommitMessages(mock.Anything, []kafka.Message{second}).Return(nil).Once()

	observer.EXPECT().ObserveOperation(OpFetch, ResultSuccess, mock.Anything).Times(2)
	observer.EXPECT().ObserveOperation(OpDecode, ResultSuccess, mock.Anything).Times(2)
	observer.EXPECT().ObserveOperation(OpHandler, ResultSuccess, mock.Anything).Times(2)
	observer.EXPECT().ObserveOperation(OpCommit, ResultError, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpCommit, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpFetch, ResultCanceled, mock.Anything).Once()

	c := newTestConsumer(t, reader, observer, noopHandle)
	require.NoError(t, c.Consume(ctx))
}

// Отмена внутри handler классифицируется как canceled, а не error.
func TestObserver_HandlerCancellation(t *testing.T) {
	reader := newMockmessageReader(t)
	observer := NewMockObserver(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			cancel()

			return validNotificationMessage(t), nil
		},
	).Once()

	handleFunc := func(ctx context.Context, _ *pb.Notification) error {
		return ctx.Err()
	}

	observer.EXPECT().ObserveOperation(OpFetch, ResultSuccess, mock.Anything).Once()
	observer.EXPECT().ObserveOperation(OpDecode, ResultSuccess, mock.Anything).Once()
	// Существующая семантика ретраит любую ошибку handler, включая отмену:
	// каждая попытка наблюдается с результатом canceled. После исчерпания
	// попыток цикл выходит по отменённому контексту без нового fetch.
	observer.EXPECT().ObserveOperation(OpHandler, ResultCanceled, mock.Anything).Times(MaxAttempts)

	c := newTestConsumer(t, reader, observer, handleFunc)
	require.NoError(t, c.Consume(ctx))
}

// Без observer потребитель сохраняет прежнее поведение и не паникует.
func TestConsume_NilObserver(t *testing.T) {
	reader := newMockmessageReader(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader.EXPECT().FetchMessage(mock.Anything).RunAndReturn(
		func(context.Context) (kafka.Message, error) {
			cancel()

			return kafka.Message{}, context.Canceled
		},
	).Once()

	c := newTestConsumer(t, reader, nil, noopHandle)
	require.NoError(t, c.Consume(ctx))
}

// Прежние вызовы конструктора без observer продолжают собираться и работать.
func TestNewConsumer_BackwardCompatible(t *testing.T) {
	logger := logmocks.NewMockLogger(t)
	logger.EXPECT().Info(mock.Anything, mock.Anything).Maybe()
	// kafka-go в фоне логирует попытки подключения к несуществующему брокеру.
	logger.EXPECT().Printf(mock.Anything, mock.Anything).Maybe()
	logger.EXPECT().Error(mock.Anything, mock.Anything).Maybe()

	opts := &ConsumerOptions{
		Brokers: []string{"localhost:9092"},
		GroupID: "observer-compat-test",
	}

	c, err := NewConsumer(
		func() *pb.Notification { return &pb.Notification{} },
		noopHandle,
		opts,
		WithTopic("observer-compat-test"),
		WithLogger(logger),
	)
	require.NoError(t, err)
	require.NotNil(t, c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, c.Consume(ctx))
	// Close останавливает фоновые горутины reader, чтобы они не логировали
	// после завершения теста.
	require.NoError(t, c.Close())
}

package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestServiceImpl_refusesANilInput sends every RPC whose request carries an Input a request
// without one.
//
// The converters dereference what they are handed, so a handler that converts before it checks
// panics, and the recovery interceptor answers the caller Internal for a request that was theirs
// to fix. It is a sweep rather than a test per handler so that an RPC added later is covered
// without anybody remembering to.
func TestServiceImpl_refusesANilInput(T *testing.T) {
	T.Parallel()

	contextType := reflect.TypeFor[context.Context]()
	errorType := reflect.TypeFor[error]()

	impl := reflect.ValueOf(buildServiceImplForTest(T))

	swept := 0
	for i := range impl.NumMethod() {
		method := impl.Type().Method(i)
		fn := method.Type

		// (receiver, ctx, *Request) (*Response, error)
		if fn.NumIn() != 3 || fn.NumOut() != 2 || fn.In(1) != contextType || fn.Out(1) != errorType || fn.In(2).Kind() != reflect.Pointer {
			continue
		}

		input, ok := fn.In(2).Elem().FieldByName("Input")
		if !ok || input.Type.Kind() != reflect.Pointer {
			continue
		}

		swept++

		T.Run(method.Name, func(t *testing.T) {
			t.Parallel()

			request := reflect.New(fn.In(2).Elem())

			// Signed in, so a handler that asks who is calling first gets as far as the input.
			ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
				Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
				ActiveAccountID: identifiers.New(),
			})

			var err error
			require.NotPanics(t, func() {
				out := impl.Method(i).Call([]reflect.Value{reflect.ValueOf(ctx), request})
				err, _ = out[1].Interface().(error)
			})

			require.Error(t, err)
			require.ErrorIs(t, err, platformerrors.ErrEmptyInputParameter)

			// And the code the error encoder answers the caller with.
			encoded := errorsgrpc.UnaryErrorEncodingInterceptor()
			_, encodedErr := encoded(t.Context(), nil, nil, func(context.Context, any) (any, error) { return nil, err })
			assert.Equal(t, codes.InvalidArgument, status.Code(encodedErr))
		})
	}

	require.Positive(T, swept, "no RPC with an Input was found to sweep")
}

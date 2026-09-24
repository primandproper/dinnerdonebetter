package grpcapi

import (
	"context"

	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"

	grpc "google.golang.org/grpc"
)

// The error encoding interceptors put a failure on the wire twice: once as a client-safe status
// message, and once as the whole wrapped chain, encoded into the status details so a trusted peer
// can reconstruct it with errorsgrpc.DecodeErrorFromStatus. That second copy is unredacted — table
// names, the rule a query broke, which of two refusals signin deliberately answers identically — and
// primitives-go is explicit that a server reachable by untrusted clients must strip it at the edge.
//
// This server is that edge. The iOS app and the web frontend dial it directly, and nothing between
// the handler and their transport removes a detail. Nor is there a trusted peer on the far side to
// keep it for: nothing that calls this server decodes the chain. So it is stripped here, for every
// method, rather than per service or behind a flag somebody has to remember to set.
//
// Only the encoded chain goes. The code, the message and the google.rpc.ErrorInfo a client branches
// on are left exactly as the encoder built them — see errorsgrpc.StripEncodedErrorDetail.

// StripEncodedErrorDetailUnaryInterceptor removes the encoded error chain from whatever status the
// interceptors inside it return. It must sit outside the error encoding interceptor, since that is
// what attaches the detail.
func StripEncodedErrorDetailUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		return resp, errorsgrpc.StripEncodedErrorDetail(err)
	}
}

// StripEncodedErrorDetailStreamInterceptor is StripEncodedErrorDetailUnaryInterceptor for streams.
func StripEncodedErrorDetailStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return errorsgrpc.StripEncodedErrorDetail(handler(srv, ss))
	}
}

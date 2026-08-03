package converters

import (
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertRequestToGRPCRequest converts a platform data privacy request to its gRPC
// representation.
//
// Failures and Retained are carried across rather than summarized. A completed
// export with entries in Failures is a partial export, and a client that renders
// "completed" over three missing sections has told a subject with thirty days to
// complain that they received everything.
func ConvertRequestToGRPCRequest(input *platformdataprivacy.Request) *dataprivacysvc.DataPrivacyRequest {
	if input == nil {
		return nil
	}

	output := &dataprivacysvc.DataPrivacyRequest{
		Id:            input.ID,
		SubjectId:     input.Subject.ID,
		RequestType:   string(input.Type),
		Status:        string(input.Status),
		RequestedAt:   timestamppb.New(input.RequestedAt),
		DueAt:         timestamppb.New(input.DueAt),
		Failures:      input.Failures,
		Retained:      input.Retained,
		Deleted:       input.Deleted,
		Anonymized:    input.Anonymized,
		ArtifactBytes: input.ArtifactBytes,
		Attempts:      int32(input.Attempts),
		LastError:     input.LastError,
	}

	// Zero rather than absent is the wrong thing to send for either of these: an
	// erasure has no expiry once confirmed, and an unfulfilled request has no
	// completion, and a client rendering the Unix epoch for both would be showing a
	// date that means "never happened".
	if !input.ExpiresAt.IsZero() {
		output.ExpiresAt = timestamppb.New(input.ExpiresAt)
	}

	if input.CompletedAt != nil {
		output.CompletedAt = timestamppb.New(*input.CompletedAt)
	}

	return output
}

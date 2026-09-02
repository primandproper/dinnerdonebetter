// Package converters carries waitlists between the wire and the platform store's
// shapes.
package converters

import (
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	waitlists "github.com/primandproper/platform-go/v13/waitlists"
)

// ConvertWaitlistToGRPCWaitlist converts a stored list to proto.
//
// The scope is deliberately not on the wire. Every list this application keeps
// is in the global scope, so a column carrying it would say the same thing on
// every row.
func ConvertWaitlistToGRPCWaitlist(input *waitlists.List) *waitlistssvc.Waitlist {
	if input == nil {
		return nil
	}

	return &waitlistssvc.Waitlist{
		Id:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		ClosesAt:      grpcconverters.ConvertTimeToPBTimestamp(input.ClosesAt),
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
	}
}

// ConvertGRPCWaitlistToWaitlist converts a proto list back to the platform's,
// for a client asserting against what it was handed.
func ConvertGRPCWaitlistToWaitlist(input *waitlistssvc.Waitlist) *waitlists.List {
	if input == nil {
		return nil
	}

	return &waitlists.List{
		ID:            input.GetId(),
		Name:          input.GetName(),
		Description:   input.GetDescription(),
		ClosesAt:      grpcconverters.ConvertPBTimestampToTime(input.GetClosesAt()),
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.GetCreatedAt()),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetLastUpdatedAt()),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.GetArchivedAt()),
	}
}

// ConvertGRPCWaitlistCreationRequestInputToWaitlist builds the list the store
// opens from what the client sent.
//
// The id and the creation time are left unset: the store mints one and reads the
// other back from the database's clock.
func ConvertGRPCWaitlistCreationRequestInputToWaitlist(input *waitlistssvc.WaitlistCreationRequestInput) *waitlists.List {
	if input == nil {
		return nil
	}

	return &waitlists.List{
		Name:        input.GetName(),
		Description: input.GetDescription(),
		ClosesAt:    grpcconverters.ConvertPBTimestampToTime(input.GetClosesAt()),
	}
}

// ConvertWaitlistToGRPCWaitlistCreationRequestInput builds the creation input a
// client would have sent to open this list. It is what the integration suite
// creates its fakes with.
func ConvertWaitlistToGRPCWaitlistCreationRequestInput(input *waitlists.List) *waitlistssvc.WaitlistCreationRequestInput {
	if input == nil {
		return nil
	}

	return &waitlistssvc.WaitlistCreationRequestInput{
		Name:        input.Name,
		Description: input.Description,
		ClosesAt:    grpcconverters.ConvertTimeToPBTimestamp(input.ClosesAt),
	}
}

// ApplyGRPCWaitlistUpdateRequestInput merges an update input into a list the
// caller has already read.
//
// It is a mutation of the read row rather than a conversion into a fresh one,
// because platform's UpdateList takes a whole List: a value built from the
// request alone would write an empty name over a list the client only wanted to
// extend.
func ApplyGRPCWaitlistUpdateRequestInput(list *waitlists.List, input *waitlistssvc.WaitlistUpdateRequestInput) {
	if list == nil || input == nil {
		return
	}

	if input.Name != nil {
		list.Name = input.GetName()
	}

	if input.Description != nil {
		list.Description = input.GetDescription()
	}

	if input.ClosesAt != nil {
		list.ClosesAt = grpcconverters.ConvertPBTimestampToTime(input.GetClosesAt())
	}
}

// ConvertWaitlistSignupToGRPCWaitlistSignup converts a stored signup to proto.
//
// The contact digest is deliberately not on the wire. It is what makes a
// withdrawal outlive the address it is about, it is of no use to a client, and
// handing it out would let anybody holding a list of addresses test which of
// them are on which list.
func ConvertWaitlistSignupToGRPCWaitlistSignup(input *waitlists.Signup) *waitlistssvc.WaitlistSignup {
	if input == nil {
		return nil
	}

	return &waitlistssvc.WaitlistSignup{
		Id:              input.ID,
		WaitlistId:      input.ListID,
		Contact:         input.Contact,
		Notes:           input.Notes,
		Status:          input.Status.String(),
		SubjectType:     input.Subject.Type.String(),
		SubjectId:       input.Subject.ID,
		CreatedAt:       grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:   grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		StatusChangedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.StatusChangedAt),
		ArchivedAt:      grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
	}
}

// ConvertGRPCWaitlistSignupToWaitlistSignup converts a proto signup back to the
// platform's, for a client asserting against what it was handed.
func ConvertGRPCWaitlistSignupToWaitlistSignup(input *waitlistssvc.WaitlistSignup) *waitlists.Signup {
	if input == nil {
		return nil
	}

	return &waitlists.Signup{
		ID:         input.GetId(),
		ListID:     input.GetWaitlistId(),
		Contact:    input.GetContact(),
		Notes:      input.GetNotes(),
		Status:     waitlists.Status(input.GetStatus()),
		Subject:    waitlists.Subject{Type: waitlists.SubjectType(input.GetSubjectType()), ID: input.GetSubjectId()},
		CreatedAt:  grpcconverters.ConvertPBTimestampToTime(input.GetCreatedAt()),
		ArchivedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetArchivedAt()),

		LastUpdatedAt:   grpcconverters.ConvertPBTimestampToTimePointer(input.GetLastUpdatedAt()),
		StatusChangedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetStatusChangedAt()),
	}
}

// ConvertGRPCWaitlistSignupCreationRequestInputToSignup builds the signup the
// store writes from what the client sent.
//
// The contact and the subject are parameters rather than fields on the input,
// because both come from the authenticated session: a signup that could name its
// own address is one anybody could make on anybody's behalf, and the address is
// what a withdrawal later suppresses. The status and the stamps are left unset —
// a signup is born waiting and the store ignores a caller that says otherwise.
func ConvertGRPCWaitlistSignupCreationRequestInputToSignup(
	input *waitlistssvc.WaitlistSignupCreationRequestInput,
	contact string,
	subject waitlists.Subject,
) *waitlists.Signup {
	if input == nil {
		return nil
	}

	return &waitlists.Signup{
		Contact: contact,
		Notes:   input.GetNotes(),
		Subject: subject,
	}
}

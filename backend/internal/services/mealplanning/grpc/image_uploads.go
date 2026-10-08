package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/uploads"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// maxImageUploadSize is the largest object these RPCs accept. It is smaller than the cap on
// platform's own upload surface (uploadedmedia.MaxUploadBytes) because everything uploaded here
// is shown inline beside a recipe, a meal or a catalog entry.
const maxImageUploadSize = 5 * 1024 * 1024 // 5 MB

// acceptedContentType is the content-type rule these RPCs apply: platform's allowlist over the
// types this application stores, which is the rule platform's own upload surface is built with
// (see internal/build/mediaregistry). A missing type is refused by the same rule, since it names
// nothing the list admits.
var acceptedContentType = mediaregistrygrpc.AllowContentTypes(uploadedmedia.SupportedMimeTypes()...)

var (
	// errBelongsToAnotherSubject refuses a header whose belongs_to names something other than the
	// subject the RPC uploads to.
	errBelongsToAnotherSubject = platformerrors.New("an upload's belongs_to must name the subject it is uploaded to")

	// errEmptyUpload refuses an upload whose stream ended without a byte of the object.
	errEmptyUpload = platformerrors.New("an upload must carry at least one byte")
)

// validObjectName reports whether a client's object name can be the last segment of a key:
// present, one segment, and not one of the two that name a directory.
//
// The key is joined from the name, so a name carrying a separator or a ".." walks out of the
// prefix the handler built — "../../../<somebody>/x.png" is an object under somebody else's
// part of the bucket. It is the rule mediaregistry/grpc applies to its own uploads, which
// platform keeps unexported; primandproper/platform-go#1163 asks for it.
func validObjectName(name string) bool {
	switch {
	case name == "", name == ".", name == "..":
		return false
	case strings.ContainsAny(name, `/\`):
		return false
	default:
		return true
	}
}

// belongsToAgrees reports whether a header's belongs_to can stand: absent, or naming the
// subject the RPC files the object under.
//
// Each of these RPCs names its subject in its own field and attaches the object to it through
// a bridge table, so that is the subject the registry records. The header is platform's, and a
// client that fills in its belongs_to is saying the same thing a second time. One that says
// something else is refused rather than overruled, so it is never told an upload succeeded
// that was filed under a subject it did not name.
func belongsToAgrees(claimed, subject mediaregistry.Subject) bool {
	return claimed == (mediaregistry.Subject{}) || claimed == subject
}

// What a piece of media hangs off in the registry's belongs-to pair. The
// vocabulary is this application's — the registry neither knows nor validates
// it — so the words are declared once rather than spelled at each write.
//
// The bridge tables are still what order media within a thing and record who
// attached it. The subject says an object is one of this recipe's, which is what
// makes an orphan sweep and a "what is attached to this" read possible without
// consulting five tables.
const (
	mealSubjectType             = "meal"
	recipeSubjectType           = "recipe"
	recipeStepSubjectType       = "recipe_step"
	validPreparationSubjectType = "valid_preparation"
	validIngredientSubjectType  = "valid_ingredient"
)

// mediaUploadRequest is what the five upload RPCs' messages share: platform's upload message —
// a header, then chunks — beside the RPC's own fields naming what the upload is for.
type mediaUploadRequest interface {
	GetUpload() *mediaregistrypb.UploadObjectRequest
}

// mediaTarget is what an upload RPC decides from its first message, once it has authorized the
// caller against the thing that message names: the subject the object is filed under, the part
// of the bucket its bytes go to, and the bridge row that attaches it.
type mediaTarget struct {
	attach  func(ctx context.Context, objectID string) error
	subject mediaregistry.Subject
	prefix  string
}

// uploadMedia is the body of every upload RPC here. The RPC supplies what differs between them —
// target, which reads the first message and authorizes it, and respond, which builds the RPC's
// own response — and this does the rest.
//
// # Why this is not mediaregistry/grpc's UploadObject
//
// Platform's surface attaches an upload to its sender or to nothing, because it cannot know
// whether recipe 123 is the caller's; an attachment to one of the consumer's nouns, its package
// documentation says, goes through the consumer's own RPC. These are those RPCs. What they take
// from platform is the policy rather than the server: the content-type allowlist, the name rule,
// the refusals a client already reads from platform's surface, and its arrangement of the work.
//
// # The order
//
// Platform's: every rule that can refuse the upload is asked before a byte of it is read — the
// name and content type first, since they cost nothing, then the subject, which costs a read. A
// refusal costs the client one message rather than the object.
//
// Then the bytes go to the UploadManager as they arrive, outside any transaction, and nothing
// buffers the object; the cap is enforced as they go past. The row is written afterwards in a
// short transaction of its own, so a slow client holds a stream open rather than a connection.
// A failed registration removes the bytes it just wrote, at a key this call minted. The bridge
// row is written last, and a failure of it fails the RPC: an image the client is told it
// attached to a recipe and that the recipe does not show is worse than a retry.
func uploadMedia[Req any, Res any, ReqPtr interface {
	*Req
	mediaUploadRequest
}](
	ctx context.Context,
	s *serviceImpl,
	stream grpc.ClientStreamingServer[Req, Res],
	logger logging.Logger,
	span tracing.Span,
	target func(first ReqPtr, userID string) (*mediaTarget, error),
	respond func(details *types.ResponseDetails, uploadedMediaID string) *Res,
) error {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.GetUserID()

	received, err := stream.Recv()
	if err != nil && !errors.Is(err, io.EOF) {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, streamBroken(err), "reading an upload's header")
	}

	first := ReqPtr(received)
	header := first.GetUpload().GetHeader()
	switch {
	case header == nil:
		return errorsgrpc.PrepareAndLogGRPCStatus(mediaregistrygrpc.ErrNoUploadHeader, logger, span, codes.InvalidArgument, "reading an upload's header")
	case !validObjectName(header.GetName()):
		return errorsgrpc.PrepareAndLogGRPCStatus(mediaregistrygrpc.ErrInvalidObjectName, logger, span, codes.InvalidArgument, "reading an upload's name")
	case !acceptedContentType(header.GetContentType()):
		return errorsgrpc.PrepareAndLogGRPCStatus(mediaregistrygrpc.ErrContentTypeRefused, logger, span, codes.InvalidArgument, "checking an upload's content type")
	}

	// The RPC's own refusals come back prepared, so they are passed through as they are.
	into, err := target(first, userID)
	if err != nil {
		return err
	}

	if !belongsToAgrees(mediaregistrygrpc.SubjectFromProto(header.GetBelongsTo()), into.subject) {
		return errorsgrpc.PrepareAndLogGRPCStatus(errBelongsToAnotherSubject, logger, span, codes.InvalidArgument, "reading an upload's subject")
	}

	body := &uploadReader{
		recv: func() (*mediaregistrypb.UploadObjectRequest, error) {
			msg, recvErr := stream.Recv()

			return ReqPtr(msg).GetUpload(), recvErr
		},
		limit: maxImageUploadSize,
	}

	// The first chunk is waited for before anything is written, so an upload that sent only its
	// header is refused rather than stored as an empty object.
	if err = body.fill(); err != nil {
		if errors.Is(err, io.EOF) {
			err = errEmptyUpload
		}

		return body.refuse(err, logger, span)
	}

	objectID := identifiers.New()
	key := path.Join(into.prefix, objectID, header.GetName())
	contentType := header.GetContentType()
	logger = logger.WithValue(mediaregistry.ObjectAttributeKey, objectID)

	occupied, err := s.uploadManager.Exists(ctx, key)
	switch {
	case err != nil:
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "checking whether an upload's key is free")
	case occupied:
		return errorsgrpc.PrepareAndLogGRPCStatus(mediaregistry.ErrObjectKeyOccupied, logger, span, codes.AlreadyExists, "checking whether an upload's key is free")
	}

	if err = s.uploadManager.Save(ctx, key, body, uploads.WithContentType(contentType)); err != nil {
		return body.refuse(err, logger, span)
	}

	created, err := inTransaction(ctx, s.db, func(tx database.Tx) (*mediaregistry.Object, error) {
		return s.registry.RecordObject(ctx, tx, tenancy.Global(), mediaregistry.ObjectInput{
			ID:          objectID,
			Key:         key,
			ContentType: contentType,
			OwnerID:     userID,
			BelongsTo:   into.subject,
			Size:        body.n,
		})
	})
	if err != nil {
		if deleteErr := s.uploadManager.Delete(ctx, key); deleteErr != nil {
			logger.Error("removing the bytes of an upload whose registration failed", deleteErr)
		}

		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "registering an upload")
	}

	if err = into.attach(ctx, created.ID); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "attaching an upload to its subject")
	}

	details := &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()}
	if err = stream.SendAndClose(respond(details, created.ID)); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "answering an upload")
	}

	logger.Info("media uploaded")

	return nil
}

// uploadReader is an io.Reader over an upload's chunks, after its header.
//
// It holds the chunk the stream last delivered and nothing else, so an upload costs one
// message's worth of memory. It counts what it receives, which is the size the row records, and
// refuses the chunk that would take the count past limit before handing any of it over, so the
// manager is never given a byte beyond the cap.
//
// err is the client's mistake, kept apart from io.EOF so the RPC can say what they did rather
// than what the manager made of it. broken is the stream's: the client went away or its deadline
// passed partway through, which the manager would otherwise report as a failure to store.
//
// It is mediaregistry/grpc's streamReader, over the platform message each of these RPCs'
// messages wraps; primandproper/platform-go#1163 asks for one this could be instead.
type uploadReader struct {
	recv   func() (*mediaregistrypb.UploadObjectRequest, error)
	err    error
	broken error
	buf    []byte
	n      int64
	limit  int64
	done   bool
}

var _ io.Reader = (*uploadReader)(nil)

func (r *uploadReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}

	copied := copy(p, r.buf)
	r.buf = r.buf[copied:]

	return copied, nil
}

// fill receives until it holds a chunk with something in it, or the stream ends.
func (r *uploadReader) fill() error {
	for len(r.buf) == 0 {
		switch {
		case r.err != nil:
			return r.err
		case r.broken != nil:
			return r.broken
		case r.done:
			return io.EOF
		}

		upload, err := r.recv()
		if errors.Is(err, io.EOF) {
			r.done = true

			return io.EOF
		}

		if err != nil {
			r.broken = err

			return err
		}

		if upload.GetHeader() != nil {
			r.err = mediaregistrygrpc.ErrRepeatedUploadHeader

			return r.err
		}

		chunk := upload.GetChunk()
		if r.n+int64(len(chunk)) > r.limit {
			r.err = mediaregistrygrpc.ErrObjectTooLarge

			return r.err
		}

		r.buf = chunk
		r.n += int64(len(chunk))
	}

	return nil
}

// refuse turns a failure to read or store the body into the RPC's answer. The reader's own
// failure is the one the client caused and is told about; the manager's wrapping of it is the
// same fact with less in it. A stream that broke is neither the client's mistake nor the
// server's fault.
func (r *uploadReader) refuse(err error, logger logging.Logger, span tracing.Span) error {
	switch {
	case r.err != nil:
		return errorsgrpc.PrepareAndLogGRPCStatus(r.err, logger, span, codes.InvalidArgument, "receiving an upload's bytes")
	case r.broken != nil:
		return errorsgrpc.PrepareAndLogGRPCStatus(r.broken, logger, span, streamBroken(r.broken), "receiving an upload's bytes")
	case errors.Is(err, errEmptyUpload):
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "receiving an upload's bytes")
	default:
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "storing an upload")
	}
}

// streamBroken is the code for a stream that failed to deliver its next message: the
// transport's own, where it names one — Canceled for a client that went away, DeadlineExceeded
// for one whose deadline passed — and Canceled otherwise. Never Internal: nothing on this side
// failed. It is mediaregistry/grpc's rule, which platform keeps unexported; see
// primandproper/platform-go#1163.
func streamBroken(err error) codes.Code {
	if st, ok := status.FromError(err); ok && st.Code() != codes.Unknown && st.Code() != codes.OK {
		return st.Code()
	}

	return codes.Canceled
}

func (s *serviceImpl) UploadMealImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadMealMediaRequest, mealplanningsvc.UploadMealImageResponse]) error {
	ctx, span := s.tracer.StartSpan(stream.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	return uploadMedia(ctx, s, stream, logger, span,
		func(first *mealplanningsvc.UploadMealMediaRequest, userID string) (*mediaTarget, error) {
			mealID := first.GetMealId()
			if mealID == "" {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("meal_id is required"), logger, span, codes.InvalidArgument, "meal_id is required")
			}

			meal, err := s.mealPlanningManager.ReadMeal(ctx, mealID)
			if err != nil || meal == nil {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(fmt.Errorf("meal not found or access denied: %w", err), logger, span, codes.PermissionDenied, "meal not found or access denied")
			}

			if meal.CreatedByUser != userID {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "permission denied")
			}

			return &mediaTarget{
				subject: mediaregistry.Subject{Type: mealSubjectType, ID: mealID},
				prefix:  path.Join("meals", mealID),
				attach: func(ctx context.Context, objectID string) error {
					return s.mealPlanningManager.AddMealImage(ctx, mealID, objectID, userID)
				},
			}, nil
		},
		func(details *types.ResponseDetails, uploadedMediaID string) *mealplanningsvc.UploadMealImageResponse {
			return &mealplanningsvc.UploadMealImageResponse{ResponseDetails: details, UploadedMediaId: &uploadedMediaID}
		},
	)
}

func (s *serviceImpl) UploadRecipeImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadRecipeMediaRequest, mealplanningsvc.UploadRecipeImageResponse]) error {
	ctx, span := s.tracer.StartSpan(stream.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	return uploadMedia(ctx, s, stream, logger, span,
		func(first *mealplanningsvc.UploadRecipeMediaRequest, userID string) (*mediaTarget, error) {
			recipeID := first.GetRecipeId()
			if recipeID == "" {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("recipe_id is required"), logger, span, codes.InvalidArgument, "recipe_id is required")
			}

			// The manager decides who may attach to a recipe; a recipe that is not the caller's is not found.
			if err := s.mealPlanningManager.AuthorizeRecipeImageUpload(ctx, recipeID, userID); err != nil {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "authorizing a recipe image upload")
			}

			return &mediaTarget{
				subject: mediaregistry.Subject{Type: recipeSubjectType, ID: recipeID},
				prefix:  path.Join("recipes", recipeID),
				attach: func(ctx context.Context, objectID string) error {
					return s.mealPlanningManager.AddRecipeImage(ctx, recipeID, objectID, userID)
				},
			}, nil
		},
		func(details *types.ResponseDetails, uploadedMediaID string) *mealplanningsvc.UploadRecipeImageResponse {
			return &mealplanningsvc.UploadRecipeImageResponse{ResponseDetails: details, UploadedMediaId: &uploadedMediaID}
		},
	)
}

func (s *serviceImpl) UploadPreparationMedia(stream grpc.ClientStreamingServer[mealplanningsvc.UploadPreparationMediaRequest, mealplanningsvc.UploadPreparationMediaResponse]) error {
	ctx, span := s.tracer.StartSpan(stream.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	return uploadMedia(ctx, s, stream, logger, span,
		func(first *mealplanningsvc.UploadPreparationMediaRequest, _ string) (*mediaTarget, error) {
			validPreparationID := first.GetValidPreparationId()
			if validPreparationID == "" {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("valid_preparation_id is required"), logger, span, codes.InvalidArgument, "valid_preparation_id is required")
			}

			if _, err := s.mealPlanningManager.ReadValidPreparation(ctx, validPreparationID); err != nil {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(fmt.Errorf("preparation not found: %w", err), logger, span, codes.NotFound, "preparation not found")
			}

			var forIngredientID *string
			if v := first.GetForIngredientId(); v != "" {
				forIngredientID = &v
			}

			return &mediaTarget{
				subject: mediaregistry.Subject{Type: validPreparationSubjectType, ID: validPreparationID},
				prefix:  path.Join("preparations", validPreparationID),
				attach: func(ctx context.Context, objectID string) error {
					return s.mealPlanningManager.AddPreparationMedia(ctx, validPreparationID, forIngredientID, objectID, 0)
				},
			}, nil
		},
		func(details *types.ResponseDetails, uploadedMediaID string) *mealplanningsvc.UploadPreparationMediaResponse {
			return &mealplanningsvc.UploadPreparationMediaResponse{ResponseDetails: details, UploadedMediaId: &uploadedMediaID}
		},
	)
}

func (s *serviceImpl) UploadIngredientMedia(stream grpc.ClientStreamingServer[mealplanningsvc.UploadIngredientMediaRequest, mealplanningsvc.UploadIngredientMediaResponse]) error {
	ctx, span := s.tracer.StartSpan(stream.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	return uploadMedia(ctx, s, stream, logger, span,
		func(first *mealplanningsvc.UploadIngredientMediaRequest, _ string) (*mediaTarget, error) {
			validIngredientID := first.GetValidIngredientId()
			if validIngredientID == "" {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("valid_ingredient_id is required"), logger, span, codes.InvalidArgument, "valid_ingredient_id is required")
			}

			if _, err := s.mealPlanningManager.ReadValidIngredient(ctx, validIngredientID); err != nil {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(fmt.Errorf("ingredient not found: %w", err), logger, span, codes.NotFound, "ingredient not found")
			}

			return &mediaTarget{
				subject: mediaregistry.Subject{Type: validIngredientSubjectType, ID: validIngredientID},
				prefix:  path.Join("ingredients", validIngredientID),
				attach: func(ctx context.Context, objectID string) error {
					return s.mealPlanningManager.AddIngredientMedia(ctx, validIngredientID, objectID, 0)
				},
			}, nil
		},
		func(details *types.ResponseDetails, uploadedMediaID string) *mealplanningsvc.UploadIngredientMediaResponse {
			return &mealplanningsvc.UploadIngredientMediaResponse{ResponseDetails: details, UploadedMediaId: &uploadedMediaID}
		},
	)
}

func (s *serviceImpl) UploadRecipeStepImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadRecipeStepImageRequest, mealplanningsvc.UploadRecipeStepImageResponse]) error {
	ctx, span := s.tracer.StartSpan(stream.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	return uploadMedia(ctx, s, stream, logger, span,
		func(first *mealplanningsvc.UploadRecipeStepImageRequest, userID string) (*mediaTarget, error) {
			recipeID, recipeStepID := first.GetRecipeId(), first.GetRecipeStepId()
			switch {
			case recipeID == "":
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("recipe_id is required"), logger, span, codes.InvalidArgument, "recipe_id is required")
			case recipeStepID == "":
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("recipe_step_id is required"), logger, span, codes.InvalidArgument, "recipe_step_id is required")
			}

			// The manager decides who may attach to a step; a recipe that is not the caller's, or a
			// step that is not the recipe's, is not found.
			if err := s.mealPlanningManager.AuthorizeRecipeStepImageUpload(ctx, recipeID, recipeStepID, userID); err != nil {
				return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "authorizing a recipe step image upload")
			}

			return &mediaTarget{
				subject: mediaregistry.Subject{Type: recipeStepSubjectType, ID: recipeStepID},
				prefix:  path.Join("recipes", recipeID, "steps", recipeStepID),
				attach: func(ctx context.Context, objectID string) error {
					return s.mealPlanningManager.AddRecipeStepImage(ctx, recipeID, recipeStepID, objectID, userID)
				},
			}, nil
		},
		func(details *types.ResponseDetails, uploadedMediaID string) *mealplanningsvc.UploadRecipeStepImageResponse {
			return &mealplanningsvc.UploadRecipeStepImageResponse{ResponseDetails: details, UploadedMediaId: &uploadedMediaID}
		},
	)
}

package reportartifacts

import (
	"context"
	"errors"
	"fmt"

	dataprivacykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"

	"github.com/primandproper/platform-go/v9/cryptography/encryption"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads"
)

const (
	o11yName = "report_artifact_store"

	// artifactExtension marks an object as an encrypted disclosure report. It is deliberately
	// not ".json": what is in the object is base64 ciphertext, and anything that reaches for it
	// expecting JSON should fail on the name rather than on the bytes.
	artifactExtension = ".json.enc"

	// legacyArtifactExtension is where reports were written before they were encrypted. Nothing
	// writes it any more and Open will not read it, but Delete still reaps it: the plaintext
	// objects that predate this package are exactly the ones most worth destroying.
	legacyArtifactExtension = ".json"
)

type (
	// Store is the only way in or out of a user data disclosure's artifact.
	//
	// There is no Download and no signed-URL equivalent on purpose — see the package doc.
	Store interface {
		// Save encrypts report and writes it as the artifact for reportID, replacing any
		// artifact already there.
		Save(ctx context.Context, reportID string, report []byte) error

		// Open reads the artifact for reportID and returns its plaintext.
		Open(ctx context.Context, reportID string) ([]byte, error)

		// Delete destroys the artifact for reportID. It is idempotent: an artifact that is
		// already gone is not an error, because the reaper's job is to guarantee absence,
		// not to observe a deletion.
		Delete(ctx context.Context, reportID string) error
	}

	store struct {
		logger        logging.Logger
		tracer        tracing.Tracer
		uploadManager uploads.UploadManager
		encDec        encryption.EncryptorDecryptor
	}
)

var _ Store = (*store)(nil)

// NewStore builds a Store over the given object storage and cipher.
func NewStore(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	uploadManager uploads.UploadManager,
	encDec encryption.EncryptorDecryptor,
) Store {
	return &store{
		logger:        logging.NewNamedLogger(logger, o11yName),
		tracer:        tracing.NewNamedTracer(tracerProvider, o11yName),
		uploadManager: uploadManager,
		encDec:        encDec,
	}
}

// Save implements Store.
func (s *store) Save(ctx context.Context, reportID string, report []byte) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if reportID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	path := artifactPath(reportID)
	logger := s.loggerFor(span, reportID, path)

	ciphertext, err := s.encDec.Encrypt(ctx, string(report))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "encrypting report artifact")
	}

	if err = uploads.SaveFile(ctx, s.uploadManager, path, []byte(ciphertext)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "saving report artifact")
	}

	return nil
}

// Open implements Store.
func (s *store) Open(ctx context.Context, reportID string) ([]byte, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if reportID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	path := artifactPath(reportID)
	logger := s.loggerFor(span, reportID, path)

	ciphertext, err := uploads.ReadFile(ctx, s.uploadManager, path)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "reading report artifact")
	}

	plaintext, err := s.encDec.Decrypt(ctx, string(ciphertext))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "decrypting report artifact")
	}

	return []byte(plaintext), nil
}

// Delete implements Store.
func (s *store) Delete(ctx context.Context, reportID string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if reportID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	// Both extensions are swept, so a bucket that still holds pre-encryption plaintext is
	// emptied of it by the same pass that expires the encrypted artifacts. A failure on one
	// path does not skip the other: whichever object can be destroyed, is.
	var errs []error
	for _, path := range []string{artifactPath(reportID), reportID + legacyArtifactExtension} {
		logger := s.loggerFor(span, reportID, path)

		exists, err := s.uploadManager.Exists(ctx, path)
		if err != nil {
			errs = append(errs, observability.PrepareAndLogError(err, logger, span, "checking for report artifact"))
			continue
		}

		if !exists {
			continue
		}

		if err = s.uploadManager.Delete(ctx, path); err != nil {
			errs = append(errs, observability.PrepareAndLogError(err, logger, span, "deleting report artifact"))
		}
	}

	return errors.Join(errs...)
}

func (s *store) loggerFor(span tracing.Span, reportID, path string) logging.Logger {
	tracing.AttachToSpan(span, dataprivacykeys.UserDataAggregationReportIDKey, reportID)

	return s.logger.
		WithValue(dataprivacykeys.UserDataAggregationReportIDKey, reportID).
		WithValue(dataprivacykeys.UserDataDisclosureArtifactPathKey, path)
}

// artifactPath is where the encrypted artifact for a report lives. It is the one place that
// knows the layout, so a rename here does not have to be chased through the writer, the reader,
// and the reaper separately.
func artifactPath(reportID string) string {
	return fmt.Sprintf("%s%s", reportID, artifactExtension)
}

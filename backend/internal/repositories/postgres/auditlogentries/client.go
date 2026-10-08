package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// Log is this application's audit log: platform's Recorder and Reader, built together over this
// application's table prefix.
//
// It holds no database handle and records nothing of its own. Writes go through the Recorder on
// the caller's transaction — in a repository, through platform's recording.Recorder, which reads
// the actor and any impersonator off the context and groups entries by chain — so an entry
// commits with the change it describes or not at all. The schema belongs to the platform as well:
// the uniqueness constraint that makes a forked chain unrepresentable is the guarantee rather than
// an incidental storage detail, so there is no generated querier here and no SQL in this package.
type Log struct {
	recorder platformaudit.Recorder
	reader   platformaudit.Reader
}

// ProvideAuditLog builds the audit log.
//
// The Recorder and Reader are built here rather than injected, so that the
// redaction policy and the table prefix are applied to every audit log in the
// process by construction. A Recorder assembled somewhere else could be assembled
// without them, and the failure would be silent in the direction that matters:
// entries written to a table nothing reads, or secrets written to a table nothing
// can edit.
//
// Pass the metrics provider. audit_chain_breaks is the instrument to alert on —
// everything else this package emits describes throughput, but a non-zero break
// count means the log has stopped being evidence.
func ProvideAuditLog(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
) (*Log, error) {
	if client == nil {
		return nil, platformaudit.ErrNilDatabaseClient
	}

	recorderOptions := []platformaudit.RecorderOption{
		platformaudit.WithRecorderTablePrefix(branding.TablePrefix),
		platformaudit.WithRecorderLogger(logging.EnsureLogger(logger)),
		platformaudit.WithRecorderTracerProvider(tracerProvider),
		platformaudit.WithRecorderMetricsProvider(metricsProvider),
	}

	// No redaction policy of this application's own. Credentials are platform's default
	// (audit.CredentialRedaction, installed by NewRecorder), and the personal free text on
	// platform's rows — a comment's body, an operator's note — is hashed by the RecordingHooks
	// that diff it. A resource type of this application's that needs one adds it here.
	recorder, err := platformaudit.NewRecorder(client.Dialect(), recorderOptions...)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building audit recorder")
	}

	reader, err := platformaudit.NewReader(
		client.Dialect(),
		platformaudit.WithReaderTablePrefix(branding.TablePrefix),
		platformaudit.WithReaderLogger(logging.EnsureLogger(logger)),
		platformaudit.WithReaderTracerProvider(tracerProvider),
		platformaudit.WithReaderMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building audit reader")
	}

	return &Log{
		recorder: recorder,
		reader:   reader,
	}, nil
}

// Reader is the reader this log built, for the surfaces that read the log directly.
//
// It is exposed rather than rebuilt because rebuilding is the failure this
// package's construction exists to prevent: a Reader assembled elsewhere could
// be assembled with a different table prefix, and would then answer "no entries"
// about a log that is full. platform's audit/grpc takes an audit.Reader, so this
// is how it gets the one whose prefix matches the Recorder's.
func (l *Log) Reader() platformaudit.Reader { return l.reader }

// Recorder is the recorder this log built, for the recording spine and the surfaces that record
// into the log directly — for the same reason Reader exists.
func (l *Log) Recorder() platformaudit.Recorder { return l.recorder }

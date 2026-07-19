package payments

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/keys"
	generated "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/payments/generated"

	"github.com/primandproper/platform-go/v5/database"
	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/tracing"
)

const (
	resourceTypeProducts = "products"
)

func (r *repository) CreateProduct(ctx context.Context, input *payments.ProductDatabaseCreationInput) (*payments.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := r.logger.Clone()
	logger = logger.WithValue(paymentskeys.ProductIDKey, input.ID)
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, input.ID)

	arg := &generated.CreateProductParams{
		ID:                    input.ID,
		Name:                  input.Name,
		Description:           input.Description,
		Kind:                  generated.ProductKind(input.Kind),
		AmountCents:           input.AmountCents,
		Currency:              input.Currency,
		BillingIntervalMonths: database.NullInt32FromInt32Pointer(input.BillingIntervalMonths),
		ExternalProductID:     database.NullStringFromString(input.ExternalProductID),
	}

	if err := r.generatedQuerier.CreateProduct(ctx, r.writeDB, arg); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating product")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, r.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeProducts,
		RelevantID:   input.ID,
		EventType:    audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	return r.GetProduct(ctx, input.ID)
}

func (r *repository) GetProduct(ctx context.Context, id string) (*payments.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(paymentskeys.ProductIDKey, id)
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, id)

	result, err := r.generatedQuerier.GetProduct(ctx, r.readDB, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching product")
	}

	return convertProductFromGenerated(result), nil
}

func (r *repository) GetProductByExternalID(ctx context.Context, externalProductID string) (*payments.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if externalProductID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("external_product_id", externalProductID)
	tracing.AttachToSpan(span, "external_product_id", externalProductID)

	result, err := r.generatedQuerier.GetProductByExternalID(ctx, r.readDB, database.NullStringFromString(externalProductID))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching product by external ID")
	}

	return convertProductFromGenerated(result), nil
}

func (r *repository) GetProducts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Product], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	params := &generated.GetProductsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	}

	results, err := r.generatedQuerier.GetProducts(ctx, r.readDB, params)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching products")
	}

	return convertProductsResult(results, filter), nil
}

func (r *repository) UpdateProduct(ctx context.Context, product *payments.Product) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if product == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.Clone()
	logger = logger.WithValue(paymentskeys.ProductIDKey, product.ID)
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, product.ID)

	arg := &generated.UpdateProductParams{
		ID:                    product.ID,
		Name:                  product.Name,
		Description:           product.Description,
		Kind:                  generated.ProductKind(product.Kind),
		AmountCents:           product.AmountCents,
		Currency:              product.Currency,
		BillingIntervalMonths: database.NullInt32FromInt32Pointer(product.BillingIntervalMonths),
		ExternalProductID:     database.NullStringFromString(product.ExternalProductID),
	}

	_, err := r.generatedQuerier.UpdateProduct(ctx, r.writeDB, arg)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating product")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, r.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeProducts,
		RelevantID:   product.ID,
		EventType:    audit.AuditLogEventTypeUpdated,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	return nil
}

func (r *repository) ArchiveProduct(ctx context.Context, id string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(paymentskeys.ProductIDKey, id)
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, id)

	_, err := r.generatedQuerier.ArchiveProduct(ctx, r.writeDB, id)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving product")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, r.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeProducts,
		RelevantID:   id,
		EventType:    audit.AuditLogEventTypeArchived,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	return nil
}

func (r *repository) ProductExists(ctx context.Context, id string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}

	result, err := r.generatedQuerier.CheckProductExistence(ctx, r.readDB, id)
	if err != nil {
		return false, observability.PrepareAndLogError(err, r.logger.Clone(), span, "checking product existence")
	}
	return result, nil
}

func convertProductFromGenerated(m *generated.Products) *payments.Product {
	if m == nil {
		return nil
	}
	return &payments.Product{
		ID:                    m.ID,
		Name:                  m.Name,
		Description:           m.Description,
		Kind:                  string(m.Kind),
		AmountCents:           m.AmountCents,
		Currency:              m.Currency,
		BillingIntervalMonths: database.Int32PointerFromNullInt32(m.BillingIntervalMonths),
		ExternalProductID:     database.StringFromNullString(m.ExternalProductID),
		CreatedAt:             m.CreatedAt,
		LastUpdatedAt:         database.TimePointerFromNullTime(m.LastUpdatedAt),
		ArchivedAt:            database.TimePointerFromNullTime(m.ArchivedAt),
	}
}

func convertProductFromRow(row *generated.GetProductsRow) *payments.Product {
	if row == nil {
		return nil
	}
	return &payments.Product{
		ID:                    row.ID,
		Name:                  row.Name,
		Description:           row.Description,
		Kind:                  string(row.Kind),
		AmountCents:           row.AmountCents,
		Currency:              row.Currency,
		BillingIntervalMonths: database.Int32PointerFromNullInt32(row.BillingIntervalMonths),
		ExternalProductID:     database.StringFromNullString(row.ExternalProductID),
		CreatedAt:             row.CreatedAt,
		LastUpdatedAt:         database.TimePointerFromNullTime(row.LastUpdatedAt),
		ArchivedAt:            database.TimePointerFromNullTime(row.ArchivedAt),
	}
}

func convertProductsResult(rows []*generated.GetProductsRow, filter *filtering.QueryFilter) *filtering.QueryFilteredResult[payments.Product] {
	data := make([]*payments.Product, 0, len(rows))
	var filteredCount, totalCount uint64
	for _, row := range rows {
		data = append(data, convertProductFromRow(row))
		filteredCount = uint64(row.FilteredCount)
		totalCount = uint64(row.TotalCount)
	}
	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(p *payments.Product) string { return p.ID },
		filter,
	)
}

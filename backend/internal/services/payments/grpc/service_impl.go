package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

// errInputRequired is a request whose input message is absent. It is this
// service's own refusal rather than the store's: the store's nil sentinels
// describe a nil row, and a nil row is what a missing input would become.
var errInputRequired = platformerrors.New("input is required")

// requester is the session every account-scoped read here is answered for.
//
// The account is the session's active one, and never the one a request names.
// GetSubscriptionsForAccountRequest and its siblings carry an account_id, and
// honoring it would let any member read another account's billing by asking;
// what somebody may see is what the account they are acting as has bought.
func (s *serviceImpl) requester(ctx context.Context, span tracing.Span) (*sessions.ContextData, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "fetching session context data")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, sessionContextData.GetUserID())
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	return sessionContextData, nil
}

// pageRequest is the session plus the page a list method was asked for.
func (s *serviceImpl) pageRequest(ctx context.Context, span tracing.Span, protoFilter *filteringpb.QueryFilter) (*sessions.ContextData, *filtering.QueryFilter, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	filter, err := filteringgrpc.FromProto(protoFilter)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	return sessionContextData, filter, nil
}

// responseDetails is the envelope every response carries.
func responseDetails(span tracing.Span, sessionContextData *sessions.ContextData) *types.ResponseDetails {
	return &types.ResponseDetails{
		TraceId:          span.SpanContext().TraceID().String(),
		CurrentAccountId: sessionContextData.GetActiveAccountID(),
	}
}

// CreateProduct adds a product to the catalog.
func (s *serviceImpl) CreateProduct(ctx context.Context, request *paymentssvc.CreateProductRequest) (*paymentssvc.CreateProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	created, err := s.billing.CreateProduct(ctx, ddbpayments.Scope(), converters.ConvertGRPCProductCreationRequestInputToProduct(request.GetInput()))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating product")
	}

	return &paymentssvc.CreateProductResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Created:         converters.ConvertProductToGRPC(created),
	}, nil
}

// GetProduct reads one product.
func (s *serviceImpl) GetProduct(ctx context.Context, request *paymentssvc.GetProductRequest) (*paymentssvc.GetProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.ProductIDKey, request.GetProductId())
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, request.GetProductId())

	product, err := s.billing.GetProduct(ctx, ddbpayments.Scope(), request.GetProductId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching product")
	}

	return &paymentssvc.GetProductResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertProductToGRPC(product),
	}, nil
}

// GetProducts pages the catalog.
func (s *serviceImpl) GetProducts(ctx context.Context, request *paymentssvc.GetProductsRequest) (*paymentssvc.GetProductsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.billing.ListProducts(ctx, ddbpayments.Scope(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching products")
	}

	results := make([]*paymentssvc.Product, 0, len(page.Data))
	for _, product := range page.Data {
		results = append(results, converters.ConvertProductToGRPC(product))
	}

	return &paymentssvc.GetProductsResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// UpdateProduct applies a patch to a product: what the request set is written,
// what it left unset is kept.
func (s *serviceImpl) UpdateProduct(ctx context.Context, request *paymentssvc.UpdateProductRequest) (*paymentssvc.UpdateProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.ProductIDKey, request.GetProductId())
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, request.GetProductId())

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	product, err := s.billing.GetProduct(ctx, ddbpayments.Scope(), request.GetProductId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching product")
	}

	converters.ApplyGRPCProductUpdateRequestInput(product, request.GetInput())

	if err = s.billing.UpdateProduct(ctx, ddbpayments.Scope(), product); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating product")
	}

	return &paymentssvc.UpdateProductResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// ArchiveProduct withdraws a product from sale.
func (s *serviceImpl) ArchiveProduct(ctx context.Context, request *paymentssvc.ArchiveProductRequest) (*paymentssvc.ArchiveProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.ProductIDKey, request.GetProductId())
	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, request.GetProductId())

	if err = s.billing.ArchiveProduct(ctx, ddbpayments.Scope(), request.GetProductId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving product")
	}

	return &paymentssvc.ArchiveProductResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// CreateSubscription opens an agreement by hand, which is what granting somebody
// a plan without a payment provider behind it looks like.
func (s *serviceImpl) CreateSubscription(ctx context.Context, request *paymentssvc.CreateSubscriptionRequest) (*paymentssvc.CreateSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	created, err := s.billing.CreateSubscription(ctx, ddbpayments.Scope(), converters.ConvertGRPCSubscriptionCreationRequestInputToSubscription(request.GetInput()))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating subscription")
	}

	return &paymentssvc.CreateSubscriptionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Created:         converters.ConvertSubscriptionToGRPC(created),
	}, nil
}

// GetSubscription reads one subscription.
func (s *serviceImpl) GetSubscription(ctx context.Context, request *paymentssvc.GetSubscriptionRequest) (*paymentssvc.GetSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())

	subscription, err := s.billing.GetSubscription(ctx, ddbpayments.Scope(), request.GetSubscriptionId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching subscription")
	}

	return &paymentssvc.GetSubscriptionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertSubscriptionToGRPC(subscription),
	}, nil
}

// GetSubscriptionsForAccount pages the active account's subscriptions, current
// and lapsed alike.
func (s *serviceImpl) GetSubscriptionsForAccount(ctx context.Context, request *paymentssvc.GetSubscriptionsForAccountRequest) (*paymentssvc.GetSubscriptionsForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.billing.ListSubscriptionsForAccount(ctx, ddbpayments.Scope(), sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching subscriptions")
	}

	results := make([]*paymentssvc.Subscription, 0, len(page.Data))
	for _, subscription := range page.Data {
		results = append(results, converters.ConvertSubscriptionToGRPC(subscription))
	}

	return &paymentssvc.GetSubscriptionsForAccountResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// UpdateSubscription applies a patch to a subscription.
func (s *serviceImpl) UpdateSubscription(ctx context.Context, request *paymentssvc.UpdateSubscriptionRequest) (*paymentssvc.UpdateSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	subscription, err := s.billing.GetSubscription(ctx, ddbpayments.Scope(), request.GetSubscriptionId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching subscription")
	}

	converters.ApplyGRPCSubscriptionUpdateRequestInput(subscription, request.GetInput())

	if err = s.billing.UpdateSubscription(ctx, ddbpayments.Scope(), subscription); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating subscription")
	}

	return &paymentssvc.UpdateSubscriptionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// ArchiveSubscription retires a subscription administratively. It is not a
// cancellation; see billing.SubscriptionStore.
func (s *serviceImpl) ArchiveSubscription(ctx context.Context, request *paymentssvc.ArchiveSubscriptionRequest) (*paymentssvc.ArchiveSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, request.GetSubscriptionId())

	if err = s.billing.ArchiveSubscription(ctx, ddbpayments.Scope(), request.GetSubscriptionId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving subscription")
	}

	return &paymentssvc.ArchiveSubscriptionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// GetPurchasesForAccount pages the active account's one-time purchases.
func (s *serviceImpl) GetPurchasesForAccount(ctx context.Context, request *paymentssvc.GetPurchasesForAccountRequest) (*paymentssvc.GetPurchasesForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.billing.ListPurchasesForAccount(ctx, ddbpayments.Scope(), sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching purchases")
	}

	results := make([]*paymentssvc.Purchase, 0, len(page.Data))
	for _, purchase := range page.Data {
		results = append(results, converters.ConvertPurchaseToGRPC(purchase))
	}

	return &paymentssvc.GetPurchasesForAccountResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// GetPaymentHistoryForAccount pages the active account's ledger, oldest first.
func (s *serviceImpl) GetPaymentHistoryForAccount(ctx context.Context, request *paymentssvc.GetPaymentHistoryForAccountRequest) (*paymentssvc.GetPaymentHistoryForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.billing.ListTransactionsForAccount(ctx, ddbpayments.Scope(), sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching payment history")
	}

	results := make([]*paymentssvc.PaymentTransaction, 0, len(page.Data))
	for _, transaction := range page.Data {
		results = append(results, converters.ConvertTransactionToGRPC(transaction))
	}

	return &paymentssvc.GetPaymentHistoryForAccountResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

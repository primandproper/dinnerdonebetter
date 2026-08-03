package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateProduct(ctx context.Context, request *paymentssvc.CreateProductRequest) (*paymentssvc.CreateProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertProductCreationRequestInputToDomain(request.Input)
	created, err := s.paymentsManager.CreateProduct(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create product")
	}

	return &paymentssvc.CreateProductResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Created:         converters.ConvertProductToGRPC(created),
	}, nil
}

func (s *serviceImpl) GetProduct(ctx context.Context, request *paymentssvc.GetProductRequest) (*paymentssvc.GetProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	product, err := s.paymentsManager.GetProduct(ctx, request.ProductId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.ProductIDKey, request.ProductId), span, codes.Internal, "failed to retrieve product")
	}

	return &paymentssvc.GetProductResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Result:          converters.ConvertProductToGRPC(product),
	}, nil
}

func (s *serviceImpl) GetProducts(ctx context.Context, request *paymentssvc.GetProductsRequest) (*paymentssvc.GetProductsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	results, err := s.paymentsManager.GetProducts(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to retrieve products")
	}

	x := &paymentssvc.GetProductsResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(results.Pagination, filter),
	}
	for _, p := range results.Data {
		x.Results = append(x.Results, converters.ConvertProductToGRPC(p))
	}
	return x, nil
}

func (s *serviceImpl) UpdateProduct(ctx context.Context, request *paymentssvc.UpdateProductRequest) (*paymentssvc.UpdateProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertProductUpdateRequestInputToDomain(request.Input)
	if err := s.paymentsManager.UpdateProduct(ctx, request.ProductId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.ProductIDKey, request.ProductId), span, codes.Internal, "failed to update product")
	}

	return &paymentssvc.UpdateProductResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
	}, nil
}

func (s *serviceImpl) ArchiveProduct(ctx context.Context, request *paymentssvc.ArchiveProductRequest) (*paymentssvc.ArchiveProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.paymentsManager.ArchiveProduct(ctx, request.ProductId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.ProductIDKey, request.ProductId), span, codes.Internal, "failed to archive product")
	}

	return &paymentssvc.ArchiveProductResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
	}, nil
}

func (s *serviceImpl) CreateSubscription(ctx context.Context, request *paymentssvc.CreateSubscriptionRequest) (*paymentssvc.CreateSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertSubscriptionCreationRequestInputToDomain(request.Input)
	created, err := s.paymentsManager.CreateSubscription(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create subscription")
	}

	return &paymentssvc.CreateSubscriptionResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Created:         converters.ConvertSubscriptionToGRPC(created),
	}, nil
}

func (s *serviceImpl) GetSubscription(ctx context.Context, request *paymentssvc.GetSubscriptionRequest) (*paymentssvc.GetSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sub, err := s.paymentsManager.GetSubscription(ctx, request.SubscriptionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.SubscriptionIDKey, request.SubscriptionId), span, codes.Internal, "failed to retrieve subscription")
	}

	return &paymentssvc.GetSubscriptionResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Result:          converters.ConvertSubscriptionToGRPC(sub),
	}, nil
}

func (s *serviceImpl) GetSubscriptionsForAccount(ctx context.Context, request *paymentssvc.GetSubscriptionsForAccountRequest) (*paymentssvc.GetSubscriptionsForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	results, err := s.paymentsManager.GetSubscriptionsForAccount(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to retrieve subscriptions")
	}

	x := &paymentssvc.GetSubscriptionsForAccountResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(results.Pagination, filter),
	}
	for _, sub := range results.Data {
		x.Results = append(x.Results, converters.ConvertSubscriptionToGRPC(sub))
	}
	return x, nil
}

func (s *serviceImpl) UpdateSubscription(ctx context.Context, request *paymentssvc.UpdateSubscriptionRequest) (*paymentssvc.UpdateSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertSubscriptionUpdateRequestInputToDomain(request.Input)
	if err := s.paymentsManager.UpdateSubscription(ctx, request.SubscriptionId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.SubscriptionIDKey, request.SubscriptionId), span, codes.Internal, "failed to update subscription")
	}

	return &paymentssvc.UpdateSubscriptionResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
	}, nil
}

func (s *serviceImpl) ArchiveSubscription(ctx context.Context, request *paymentssvc.ArchiveSubscriptionRequest) (*paymentssvc.ArchiveSubscriptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.paymentsManager.ArchiveSubscription(ctx, request.SubscriptionId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithValue(paymentskeys.SubscriptionIDKey, request.SubscriptionId), span, codes.Internal, "failed to archive subscription")
	}

	return &paymentssvc.ArchiveSubscriptionResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
	}, nil
}

func (s *serviceImpl) GetPurchasesForAccount(ctx context.Context, request *paymentssvc.GetPurchasesForAccountRequest) (*paymentssvc.GetPurchasesForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	results, err := s.paymentsManager.GetPurchasesForAccount(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to retrieve purchases")
	}

	x := &paymentssvc.GetPurchasesForAccountResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(results.Pagination, filter),
	}
	for _, p := range results.Data {
		x.Results = append(x.Results, converters.ConvertPurchaseToGRPC(p))
	}
	return x, nil
}

func (s *serviceImpl) GetPaymentHistoryForAccount(ctx context.Context, request *paymentssvc.GetPaymentHistoryForAccountRequest) (*paymentssvc.GetPaymentHistoryForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	results, err := s.paymentsManager.GetPaymentTransactionsForAccount(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to retrieve payment history")
	}

	x := &paymentssvc.GetPaymentHistoryForAccountResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(results.Pagination, filter),
	}
	for _, t := range results.Data {
		x.Results = append(x.Results, converters.ConvertPaymentTransactionToGRPC(t))
	}
	return x, nil
}

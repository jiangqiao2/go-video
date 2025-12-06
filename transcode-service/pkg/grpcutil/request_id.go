package grpcutil

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"transcode-service/pkg/logger"
)

const requestIDMetadataKey = "x-request-id"
const RequestIDContextKey = "request_id"

type ctxKey string

const requestIDContextKey ctxKey = "request_id"

// ContextWithRequestID ensures the context carries a request_id value and outgoing metadata.
func ContextWithRequestID(ctx context.Context, requestID string) (context.Context, string) {
	reqID := strings.TrimSpace(requestID)
	if reqID == "" {
		reqID = uuid.NewString()
	}
	ctx = context.WithValue(ctx, requestIDContextKey, reqID)
	ctx = context.WithValue(ctx, RequestIDContextKey, reqID)
	ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, reqID)
	return ctx, reqID
}

// RequestIDFromContext returns request_id from context or incoming metadata.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDContextKey).(string); ok && v != "" {
		return v
	}
	if v, ok := ctx.Value(RequestIDContextKey).(string); ok && v != "" {
		return v
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(requestIDMetadataKey); len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

// UnaryClientRequestIDInterceptor injects request_id into outgoing gRPC calls.
func UnaryClientRequestIDInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	ctx, reqID := ContextWithRequestID(ctx, RequestIDFromContext(ctx))
	err := invoker(ctx, method, req, reply, cc, opts...)
	fields := map[string]interface{}{
		"request_id":  reqID,
		"method":      method,
		"kind":        "grpc_client",
		"duration_ms": time.Since(start).Milliseconds(),
	}
	if err != nil {
		fields["error"] = err.Error()
		logger.WithFields(fields).Warn("grpc client call failed")
	} else {
		logger.WithFields(fields).Info("grpc client call")
	}
	return err
}

// UnaryServerRequestIDInterceptor extracts request_id from incoming metadata, sets it in context, and returns it to the caller.
func UnaryServerRequestIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	ctx, reqID := ContextWithRequestID(ctx, RequestIDFromContext(ctx))
	_ = grpc.SetHeader(ctx, metadata.Pairs(requestIDMetadataKey, reqID))
	resp, err := handler(ctx, req)
	fields := map[string]interface{}{
		"request_id":  reqID,
		"method":      info.FullMethod,
		"kind":        "grpc_server",
		"duration_ms": time.Since(start).Milliseconds(),
	}
	if err != nil {
		fields["error"] = err.Error()
		logger.WithFields(fields).Warn("grpc server call failed")
	} else {
		logger.WithFields(fields).Info("grpc server call")
	}
	return resp, err
}

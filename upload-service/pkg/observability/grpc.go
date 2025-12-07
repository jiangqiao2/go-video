package observability

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"upload-service/pkg/grpcutil"
)

// GRPCServerTracingInterceptor starts a span for each incoming gRPC call.
func GRPCServerTracingInterceptor(service string) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(service + "-grpc-server")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier{md: md})
		rpcService, rpcMethod := parseGRPCMethod(info.FullMethod)

		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", rpcService),
			attribute.String("rpc.method", rpcMethod),
		)
		if reqID := grpcutil.RequestIDFromContext(ctx); reqID != "" {
			span.SetAttributes(attribute.String("request.id", reqID))
		}

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		span.End()

		return resp, err
	}
}

// GRPCClientTracingInterceptor injects trace context into outgoing gRPC calls and records spans.
func GRPCClientTracingInterceptor(service string) grpc.UnaryClientInterceptor {
	tracer := otel.Tracer(service + "-grpc-client")
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		rpcService, rpcMethod := parseGRPCMethod(method)
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", rpcService),
			attribute.String("rpc.method", rpcMethod),
		)
		if reqID := grpcutil.RequestIDFromContext(ctx); reqID != "" {
			span.SetAttributes(attribute.String("request.id", reqID))
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		carrier := metadataCarrier{md: md}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, md)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		span.End()
		return err
	}
}

type metadataCarrier struct {
	md metadata.MD
}

func (c metadataCarrier) Get(key string) string {
	if len(c.md) == 0 {
		return ""
	}
	values := c.md.Get(strings.ToLower(key))
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataCarrier) Set(key, value string) {
	key = strings.ToLower(key)
	if c.md == nil {
		c.md = metadata.New(nil)
	}
	c.md.Set(key, value)
}

func (c metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

func parseGRPCMethod(fullMethod string) (string, string) {
	method := strings.TrimPrefix(fullMethod, "/")
	parts := strings.Split(method, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "unknown", method
}

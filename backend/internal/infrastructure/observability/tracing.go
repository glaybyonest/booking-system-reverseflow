package observability

import "context"

// StartSpan is a small OpenTelemetry hook point. The MVP keeps tracing optional:
// production deployments can replace this adapter with a real tracer provider
// without changing application use cases.
func StartSpan(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}

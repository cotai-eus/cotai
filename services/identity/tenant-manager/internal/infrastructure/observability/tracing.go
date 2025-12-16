package observability

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

// InitTracer initializes Jaeger tracer
func InitTracer(serviceName, agentHost string, agentPort int, logger *zap.Logger) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "probabilistic",
			Param: 0.1, // Sample 10% of traces
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: fmt.Sprintf("%s:%d", agentHost, agentPort),
		},
	}

	tracer, closer, err := cfg.NewTracer(
		config.Logger(jaegerLogger{logger: logger}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)
	logger.Info("Jaeger tracer initialized",
		zap.String("service", serviceName),
		zap.String("agent", fmt.Sprintf("%s:%d", agentHost, agentPort)),
	)

	return tracer, closer, nil
}

// jaegerLogger adapts zap.Logger to jaeger.Logger interface
type jaegerLogger struct {
	logger *zap.Logger
}

func (l jaegerLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l jaegerLogger) Infof(msg string, args ...interface{}) {
	l.logger.Sugar().Infof(msg, args...)
}

// TracedSpan creates a child span from context (if parent span exists)
func TracedSpan(operationName string) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	return tracer.StartSpan(operationName)
}

// TracedSpanFromContext creates a child span from a parent span in context
func TracedSpanFromContext(parent opentracing.SpanContext, operationName string) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	return tracer.StartSpan(
		operationName,
		opentracing.ChildOf(parent),
	)
}

// SetSpanTags sets common tags on a span
func SetSpanTags(span opentracing.Span, tags map[string]interface{}) {
	for key, value := range tags {
		span.SetTag(key, value)
	}
}

// SetSpanError marks span as error and logs the error
func SetSpanError(span opentracing.Span, err error) {
	span.SetTag("error", true)
	span.LogKV("error.message", err.Error())
}

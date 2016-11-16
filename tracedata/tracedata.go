package tracedata

import (
	"fmt"
	"math/rand"
	"net/http"

	"golang.org/x/net/context"
)

const (
	TraceIDHeader = "X-Trace-ID"

	SpanIDHeader = "X-Span-ID"

	ParentSpanIDHeader = "X-Parent-Span-ID"
)

type dataCtxKeyType int

const dataCtxKey dataCtxKeyType = 0

type TraceData struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
}

// TraceFromContext returns data from request
func GetTraceDataFromRequest(req *http.Request) *TraceData {
	traceID := req.Header.Get(TraceIDHeader)
	spanID := req.Header.Get(SpanIDHeader)
	parentSpanID := req.Header.Get(ParentSpanIDHeader)

	if traceID == "" {
		traceID = generate()
		spanID = traceID
	} else if spanID == "" {
		spanID = generate()
	}

	return &TraceData{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	}
}

// SetTraceDataToRequest sets data to request
func SetTraceDataToRequest(data *TraceData, req *http.Request) {
	req.Header.Set(TraceIDHeader, data.TraceID)
	req.Header.Set(SpanIDHeader, "") // if we do not send spanID then child service will generate spanID
	req.Header.Set(ParentSpanIDHeader, data.SpanID)
}

// SetTraceDataToContext creates new child context with given trace data
func SetTraceDataToContext(ctx context.Context, data *TraceData) context.Context {
	return context.WithValue(ctx, dataCtxKey, data)
}

// GetTraceDataFromContext gets trace data from given context
func GetTraceDataFromContext(ctx context.Context) *TraceData {
	if result, ok := ctx.Value(dataCtxKey).(*TraceData); ok {
		return result
	}
	return &TraceData{}
}

func generate() string {
	return fmt.Sprintf("%X", rand.Int63())
}

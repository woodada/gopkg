package ctxspan

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"go.opentelemetry.io/otel/trace"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Tracer struct {
}

func NewTracer() *Tracer {
	return &Tracer{}
}
func (t *Tracer) Start(ctx context.Context) context.Context {
	return FillSpanContext(ctx)
}

func (t *Tracer) Finish(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()
}

var _startTime = uint32(time.Now().Unix())
var _pid = uint32(os.Getpid()) // 16个字符, 192.168.110.111 =15
var _spanId = uint64(time.Now().UnixNano())

func GetSpanStringFromContext(ctx context.Context) string {
	return GetTraceIDFromContext(ctx) + ":" + GetSpanIDFromContext(ctx)
}

func GetTraceIDAndSpanIDFromSpanString(str string) (string, string) {
	arr := strings.Split(str, ":")
	if len(arr) == 2 {
		return arr[0], arr[1]
	}
	return "", ""
}

// ValidTraceID 校验链路跟踪字符串是否正确
func ValidTraceID(traceID string) bool {
	_, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		// fmt.Println("ERROR ValidTraceID", traceID, err)
	}
	return err == nil
}

func GetTraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func GetSpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SetSpanContextWithId 强行设置上下文里的链路标识
func SetSpanContextWithId(ctx context.Context, traceIDStr, spanIDStr string) context.Context {
	traceID, _ := trace.TraceIDFromHex(traceIDStr)
	spanID, _ := trace.SpanIDFromHex(spanIDStr)
	traceState, _ := trace.ParseTraceState("")

	return trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		TraceState: traceState,
		Remote:     false,
	}))
}

// FillSpanContextWithTraceId 兼容性的设置上下文里的链路标识 如果上下文里有链路标识则不设置
func FillSpanContextWithTraceId(ctx context.Context, traceId string) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return ctx
	}

	if !ValidTraceID(traceId) {
		return ctx
	}

	var b1 [8]byte
	binary.BigEndian.PutUint64(b1[:], atomic.AddUint64(&_spanId, 1))

	spanID := hex.EncodeToString(b1[:])

	return SetSpanContextWithId(ctx, traceId, spanID)
}

func FillSpanContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return ctx
	}

	var b2 [16]byte

	binary.BigEndian.PutUint32(b2[0:4], _pid)
	binary.BigEndian.PutUint32(b2[4:8], _startTime)
	binary.BigEndian.PutUint64(b2[8:], atomic.AddUint64(&_spanId, 1))

	traceID := hex.EncodeToString(b2[:])

	return FillSpanContextWithTraceId(ctx, traceID)
}

// UpdateSpanId 只更新SpanId
func UpdateSpanId(ctx context.Context) context.Context {
	traceID := GetTraceIDFromContext(ctx)
	if traceID == "" {
		// fmt.Println("empty")
		return ctx
	}

	var b1 [8]byte
	binary.BigEndian.PutUint64(b1[:], atomic.AddUint64(&_spanId, 1))
	spanID := hex.EncodeToString(b1[:])

	return SetSpanContextWithId(ctx, traceID, spanID)
}

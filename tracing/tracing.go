package tracing

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/no-mole/neptune/json"

	"github.com/no-mole/neptune/snowflake"
)

var TracingContextKey = "tracingContextKey"

type Tracing struct {
	Id             string    `json:"id"`
	SpanId         string    `json:"spanId"`
	StepSize       int       `json:"stepSize"`
	StartTime      time.Time `json:"startTime"`
	ParentSpanId   string    `json:"parent_span_id"`
	ParentStepSize int       `json:"parent_step_size"`
}

func Start(ctx context.Context, spanId string) context.Context {
	ctxTracing := FromContextOrNew(ctx)
	spanTracing := &Tracing{
		Id:             ctxTracing.Id,
		StepSize:       ctxTracing.StepSize + 1,
		SpanId:         spanId,
		StartTime:      time.Now(),
		ParentSpanId:   ctxTracing.SpanId,
		ParentStepSize: ctxTracing.StepSize,
	}
	return WithContext(ctx, spanTracing)
}

func WithContext(ctx context.Context, t *Tracing) context.Context {
	return context.WithValue(ctx, TracingContextKey, t)
}

func FromContextOrNew(ctx context.Context) *Tracing {
	ctxTracing := ctx.Value(TracingContextKey)
	if t, ok := ctxTracing.(*Tracing); ok {
		return t
	}
	return &Tracing{
		Id:             snowflake.GenInt64String(),
		SpanId:         "default",
		StepSize:       0,
		StartTime:      time.Now(),
		ParentStepSize: 0,
		ParentSpanId:   "background",
	}
}

func Encoding(t *Tracing) string {
	data, _ := json.Marshal(t)
	return base64.StdEncoding.EncodeToString(data)
}

func Decoding(str string) (*Tracing, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	var t *Tracing
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

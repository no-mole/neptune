package logger

import (
	"context"
	"time"

	"github.com/no-mole/neptune/tracing"

	"github.com/no-mole/neptune/logger/entry"
)

var TimeFormat = "2006-01-02T15:04:05.999Z07:00"

var TracingHandle = Handle(func(ctx context.Context, e entry.Entry) []entry.Field {
	trac := tracing.FromContextOrNew(ctx)
	return []entry.Field{
		WithField("tracingId", trac.Id),                                         //string
		WithField("tracingSpanId", trac.SpanId),                                 //string text|keyword
		WithField("tracingStepSize", trac.StepSize),                             //int
		WithField("tracingStartTime", trac.StartTime.Format(TimeFormat)),        //datetime
		WithField("tracingEndTime", time.Now().Format(TimeFormat)),              //datetime
		WithField("tracingParentSpanId", trac.ParentSpanId),                     //string
		WithField("tracingParentStepSize", trac.ParentStepSize),                 //string
		WithField("tracingDuration", time.Since(trac.StartTime).Milliseconds()), //int
	}
})

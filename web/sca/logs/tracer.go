/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	"context"

	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	jconfig "github.com/uber/jaeger-client-go/config"
	jlog "github.com/uber/jaeger-client-go/log"
)

var (
	tracers = sync.Map{}
)

// Tracing is a wrapper to opentracing tracer
type Tracing struct {
	opentracing.Tracer
}

// StartTracing new a tracing with name if it does not exist,
// or returns the existing one.
func StartTracing(name string) (tr *Tracing) {
	if v, ok := tracers.Load(name); ok {
		if tr, ok = v.(*Tracing); ok {
			return
		}
	}
	tr = &Tracing{
		newJaegerTracer(name),
	}
	tracers.Store(name, tr)
	return
}

func newJaegerTracer(service string) opentracing.Tracer {
	jaegerURI := GetJaegerURI()
	if jaegerURI == "" {
		return nil
	}
	cfg := jconfig.Configuration{
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1},
		Reporter: &jconfig.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: GetJaegerURI(),
		},
	}

	jLogger := jlog.StdLogger
	// Initialize tracer with a logger and a metrics factory
	tracer, _, err := cfg.New(
		service,
		jconfig.ZipkinSharedRPCSpan(true),
		jconfig.Logger(jLogger),
	)
	if err != nil {
		return nil
	}
	return tracer
}

// StartLogging new a span logger and put it in context,
// and return the span logger and the context.
func (tracer *Tracing) StartLogging(
	ctx context.Context,
	op string,
	opts ...opentracing.StartSpanOption) (*SpanLogger, context.Context) {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}
	sp := tracer.StartSpan(op, opts...)
	ctx = opentracing.ContextWithSpan(ctx, sp)
	logger := &SpanLogger{
		Span: sp,
		entry: logrus.NewEntry(
			defaultLogger).WithFields(
			defaultFields)}
	return logger, ctx
}

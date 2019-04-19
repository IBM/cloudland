package model

import (
	context "context"

	"github.com/IBM/cloudland/web/sca/logs"
)

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("Model")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}

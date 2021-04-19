package middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Trace() iris.Handler {
	return func(c iris.Context) {

		var parentSpan opentracing.Span

		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request().Header))
		if err != nil {
			parentSpan = opentracing.GlobalTracer().StartSpan(
				c.Request().URL.Path,
			)
		} else {
			parentSpan = opentracing.GlobalTracer().StartSpan(
				"call grpc",
				opentracing.ChildOf(spCtx),
				ext.SpanKindRPCServer,
			)
		}

		// 重置request ctx
		ctx := opentracing.ContextWithSpan(c.Request().Context(), parentSpan)
		c.ResetRequest(c.Request().WithContext(ctx))
		defer parentSpan.Finish()

		c.Next()

	}
}

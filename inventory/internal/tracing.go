package internal

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/wavefronthq/wavefront-opentracing-sdk-go/reporter"
	wfTracer "github.com/wavefronthq/wavefront-opentracing-sdk-go/tracer"

	"github.com/opentracing/opentracing-go"
	otrext "github.com/opentracing/opentracing-go/ext"
)

func NewGlobalTracer(serviceName string) io.Closer {
	config := &senders.DirectConfiguration{
		Server: "https://tracing.wavefront.com",
		Token: "354b6d73-5536-419e-be6e-779a016eeab9",
	}
	sender, err := senders.NewDirectSender(config)
	if err != nil {
		log.Fatalf("error creating wavefront sender: %q", err)
	}

	appTags := application.New("inventory", serviceName)

	directReporter := reporter.New(sender, appTags)
	consoleReporter := reporter.NewConsoleSpanReporter(serviceName)

	compositeReporter := reporter.NewCompositeSpanReporter(directReporter, consoleReporter)
	wavefrontTracer := wfTracer.New(compositeReporter)
	opentracing.SetGlobalTracer(wavefrontTracer)
	return ioutil.NopCloser(nil)
}

func NewServerSpan(req *http.Request, spanName string) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	parentCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	var span opentracing.Span
	if err == nil { // has parent context
		span = tracer.StartSpan(spanName, opentracing.ChildOf(parentCtx))
	} else if err == opentracing.ErrSpanContextNotFound { // no parent
		span = tracer.StartSpan(spanName)
	} else {
		log.Printf("Error in extracting tracer context: %s", err.Error())
	}
	otrext.SpanKindRPCServer.Set(span)
	return span
}
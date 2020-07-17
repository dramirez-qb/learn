package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

var (
	tracer opentracing.Tracer
	closer io.Closer
)

func withTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		defer log.Printf("[%s] %q", request.Method, request.URL.String())
		// log.Printf("Tracing request for %s", request.RequestURI)
		next.ServeHTTP(response, request)
	}
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log.Printf("Logged connection from %s", request.RemoteAddr)
		next.ServeHTTP(response, request)
	}
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return recoverHandler(h)
}

func recoverHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(response, http.StatusText(500), 500)
			}
		}()

		next.ServeHTTP(response, request)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("healthchek")
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	fmt.Fprintf(w, `{"alive": true}`)

	span.Finish()
}

func pongHandler(response http.ResponseWriter, request *http.Request) {
	span := tracer.StartSpan("pong")
	fmt.Fprintf(response, "pong")
	span.Finish()
}

func helloHandler(response http.ResponseWriter, request *http.Request) {
	span := tracer.StartSpan("hello")
	fmt.Fprintf(response, "Hello, %s!", request.URL.Path[1:])
	span.Finish()
}

func main() {
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return
	}
	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err = cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)

	http.Handle("/", use(helloHandler, withLogging, withTracing))
	http.Handle("/ping", use(pongHandler, withLogging, withTracing))
	http.Handle("/healthz", use(healthCheckHandler))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
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
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	fmt.Fprintf(w, `{"alive": true}`)
}

func pongHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "pong")
}

func helloHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "Hello, %s!", request.URL.Path[1:])
}

func main() {
	useTor := flag.Bool("tor", false, "use tor service")
	port := flag.String("port", "8080", "port to use")
	flag.Parse()
	http.Handle("/", use(helloHandler, withLogging, withTracing))
	http.Handle("/ping", use(pongHandler, withLogging, withTracing))
	http.Handle("/healthz", use(healthCheckHandler))
	if *useTor {
		// Start tor with some defaults + elevated verbosity
		fmt.Println("Starting and registering onion service, please wait a bit...")
		t, err := tor.Start(context.TODO(), &tor.StartConf{ProcessCreator: libtor.Creator, DebugWriter: os.Stderr})
		if err != nil {
			log.Panicf("Failed to start tor: %v", err)
		}
		defer t.Close()

		// Wait at most a few minutes to publish the service
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// Create an onion service to listen on any port but show as 80
		onion, err := t.Listen(ctx, &tor.ListenConf{RemotePorts: []int{80, 8080}})
		if err != nil {
			log.Panicf("Failed to create onion service: %v", err)
		}
		defer onion.Close()

		fmt.Printf("Please open a Tor capable browser and navigate to http://%v.onion\n", onion.ID)

		if err := http.Serve(onion, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("starting listening on %s\n", *port)
		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.Fatal(err)
		}
	}

}

package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Welcome struct {
	Name string
	Time string
	User string
}

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
	welcome := Welcome{request.URL.Path[1:], time.Now().Format(time.Stamp), os.Getenv("USER")}
	templates := template.Must(template.ParseFiles("templates/index.html"))
	if err := templates.ExecuteTemplate(response, "index.html", welcome); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func fibHandler(response http.ResponseWriter, request *http.Request) {
	num := rand.Intn(45)
	log.Printf("random: %+v", num)
	fmt.Fprintf(response, "%d\n", FibonacciRecursion(num))
}

func FibonacciRecursion(n int) int {
	if n <= 1 {
		return n
	}
	return FibonacciRecursion(n-1) + FibonacciRecursion(n-2)
}

// content is our static web server content.
//go:embed static/* templates
var content embed.FS
var gitCommit string

func main() {
	version := *flag.Bool("version", false, "Version")
	port := *flag.String("port", "8080", "port to use")
	flag.Parse()
	if version {
		fmt.Printf("version: %s\n", gitCommit)
		return
	}
	rand.Seed(time.Now().UnixNano())
	http.Handle("/", use(helloHandler, withLogging, withTracing))
	http.Handle("/fib", use(fibHandler, withLogging, withTracing))
	http.Handle("/healthz", use(healthCheckHandler))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/ping", use(pongHandler, withLogging, withTracing))
	http.Handle("/static/", http.FileServer(http.FS(content)))
	log.Printf("starting listening on %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

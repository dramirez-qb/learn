package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sys/unix"
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
	rand.Seed(time.Now().UnixNano())
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
//
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
	http.Handle("/", use(helloHandler, withLogging, withTracing))
	http.Handle("/fib", use(fibHandler, withLogging, withTracing))
	http.Handle("/healthz", use(healthCheckHandler))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/ping", use(pongHandler, withLogging, withTracing))
	http.Handle("/static/", http.FileServer(http.FS(content)))
	log.Printf("starting listening on %s\n", port)
	lc := net.ListenConfig{
		Control: control,
	}
	l, err := lc.Listen(context.TODO(), "tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	exitCh := make(chan os.Signal, 1)
	server := &http.Server{Addr: l.Addr().String()}

	signal.Notify(exitCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer func() { exitCh <- syscall.SIGTERM }()
		if err := http.Serve(l, nil); err != nil {
			fmt.Println(err)
		}
	}()
	<-exitCh
	server.Shutdown(context.Background()) // Shutdown the server
}

func control(network, address string, c syscall.RawConn) error {
	var err error
	c.Control(func(fd uintptr) {
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err != nil {
			return
		}

		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		if err != nil {
			return
		}
	})
	return err
}

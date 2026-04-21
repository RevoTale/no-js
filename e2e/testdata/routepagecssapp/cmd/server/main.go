package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gen "example.com/no-js-e2e/routepagecssapp/web/generated"
	runtime "example.com/no-js-e2e/routepagecssapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	appContext := &runtime.Context{}
	bundle := gen.Bundle(appContext)
	bundle.TemplCSSClasses = gen.TemplCSSClasses

	handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
		App: bundle,
	})
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("LISTEN_URL=http://%s\n", listener.Addr().String())

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatal(serveErr)
		}
	}()

	shutdownOnSignal(server)
}

func shutdownOnSignal(server *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Print(err)
	}
}

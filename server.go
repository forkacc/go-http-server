package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	// From Go1.16 we can use signal.NotifyContext
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		select {
		case <-gctx.Done():
			fmt.Println("Closing singal goroutine")
			return gctx.Err()
		}
		return nil
	})

	g.Go(func() error {
		return service(gctx, "0.0.0.0:8080", func(resp http.ResponseWriter, req *http.Request) {
			fmt.Fprintln(resp, "Hello World")
		})
	})

	if err := g.Wait(); err == nil || err == context.Canceled {
		fmt.Println("Finished clean")
	} else {
		fmt.Printf("Received error: %v", err)
	}
}

func service(ctx context.Context, addr string, handler http.Handler) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()
}

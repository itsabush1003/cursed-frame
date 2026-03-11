package infra

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) ListenAndServe() error {
	connClosedNotifier := make(chan struct{})

	go func() {
		c := make(chan os.Signal, 1)
		defer close(c)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(c)
		select {
		case <-c:
			break
		case <-connClosedNotifier:
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			close(connClosedNotifier)
			return
		}

		close(connClosedNotifier)
	}()

	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		close(connClosedNotifier)
		return err
	}
	<-connClosedNotifier
	return nil
}

func NewServer(addr string, handler http.Handler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	server := &Server{
		httpServer: &http.Server{
			Addr:        addr,
			Handler:     handler,
			BaseContext: func(l net.Listener) context.Context { return ctx },
		},
	}
	server.httpServer.RegisterOnShutdown(cancel)
	return server
}

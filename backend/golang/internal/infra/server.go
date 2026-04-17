package infra

import (
	"context"
	"crypto/tls"
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
	useTLS     bool
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

	listenAndServeFunc := s.httpServer.ListenAndServe
	if s.useTLS {
		listenAndServeFunc = func() error {
			return s.httpServer.ListenAndServeTLS("", "")
		}
	}

	if err := listenAndServeFunc(); !errors.Is(err, http.ErrServerClosed) {
		close(connClosedNotifier)
		return err
	}
	<-connClosedNotifier
	return nil
}

func NewServer(addr string, tlsConfig *tls.Config, handler http.Handler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	protocol := &http.Protocols{}
	protocol.SetHTTP1(true)
	protocol.SetUnencryptedHTTP2(true)
	if tlsConfig != nil && (len(tlsConfig.Certificates) > 0 || tlsConfig.GetCertificate != nil) {
		protocol.SetHTTP2(true)
	}
	server := &Server{
		httpServer: &http.Server{
			Addr:        addr,
			Handler:     handler,
			Protocols:   protocol,
			TLSConfig:   tlsConfig,
			BaseContext: func(l net.Listener) context.Context { return ctx },	//nolint: unused
		},
		useTLS: protocol.HTTP2(),
	}
	server.httpServer.RegisterOnShutdown(cancel)
	return server
}

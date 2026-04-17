package infra

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func CreateCertificateFromFiles(certFile, keyFile string) ([]tls.Certificate, error) {
	// ContainsFuncでどちらかのファイルが存在しなかった場合にtrueを返す
	if slices.ContainsFunc([]string{certFile, keyFile}, func(file string) bool {
		info, err := os.Stat(file)
		if err != nil {
			return os.IsNotExist(err)
		}
		return info.IsDir()
	}) {
		return nil, errors.New("cert file or key file is not exists.")
	}
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{certificate}, nil
}

func CreateCertificateWithAutoCert(domain string) (func(*tls.ClientHelloInfo) (*tls.Certificate, error), context.CancelFunc, error) {
	// acmeチャレンジに必要な80番ポートが使えるか確認
	listener, err := net.Listen("tcp", ":http")
	if err != nil {
		return nil, nil, err
	}
	_ = listener.Close()
	if domain == "" {
		fmt.Println("Warning: You enabled the autocert flag without specifying a domain. In this case, the domain is inferred from the first access to the server")
	} else if domain == "localhost" || strings.HasPrefix(domain, "localhost:") {
		return nil, nil, errors.New("'localhost' domain cannot create certificate")
	} else if strings.HasPrefix(domain, "*") {
		return nil, nil, errors.New("Wildcard certificate does not supported")
	}
	var mu sync.Mutex
	hostPolicy := autocert.HostWhitelist(domain)
	manager := &autocert.Manager{
		Cache:  autocert.DirCache("cert"),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if domain != "" {
				return hostPolicy(ctx, host)
			}
			mu.Lock()
			defer mu.Unlock()
			if domain == "" {
				// 'localhost'を利用したアクセスもしくはIP直打ちのアクセスの場合はエラーで弾く
				if host == "localhost" || strings.HasPrefix(host, "localhost:") || net.ParseIP(host) != nil {
					return errors.New("Invalid access")
				}
				hostPolicy = autocert.HostWhitelist(host)
				domain = host
				return nil
			}
			return hostPolicy(ctx, host)
		},
	}
	acmeServer := &http.Server{
		Addr:    ":http",
		Handler: manager.HTTPHandler(nil),
	}
	ctx, cancel := context.WithCancel(context.Background())
	// acmeチャレンジ用サーバ起動
	go acmeServer.ListenAndServe()	//nolint:errcheck
	// acmeチャレンジ用サーバをShutdownするためのgoroutine
	// 証明書が取得される前に別の理由でプログラムが終了した時に対応するため
	go func() {
		<-ctx.Done()
		shutdownCtx, forceShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer forceShutdown()
		if err := acmeServer.Shutdown(shutdownCtx); err != nil {
			_ = acmeServer.Close()
		}
	}()
	getCertificate := func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := manager.GetCertificate(hello)
		if err != nil {
			return nil, err
		}
		cancel()
		return cert, nil
	}
	return getCertificate, cancel, nil
}

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"k8s.io/klog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/petrkotas/k8s-object-lock/pkg/server"
)

var (
	// commandline params to load certificate
	paramKeyFile, paramCertFile string
	kubeconfig, masterURL       string
	port                        int
)

func main() {
	// setup kubernetes link

	serverConf := server.MakeServerConf(masterURL, kubeconfig)

	// handle tls
	pair, err := tls.LoadX509KeyPair(paramCertFile, paramKeyFile)
	if err != nil {
		klog.Errorf("Cannot load certificates: %v", err)
	}

	tlsCfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{pair},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// new server with single route for validation
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", serverConf.Validate)

	strPort := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:         strPort,
		Handler:      mux,
		TLSConfig:    tlsCfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	klog.Info("Server running")
	klog.Fatal(srv.ListenAndServeTLS("", ""))

	// allow for shutting down the server via signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}

func init() {
	klog.Info("Started validating server")
	flag.Set("v", "9")

	// parse config
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&port, "port", 443, "Default port to start server on")
	flag.StringVar(&paramCertFile, "tlsCertFile", "/etc/lockvalidation/cert/cert.pem", "File containing tls cert")
	flag.StringVar(&paramKeyFile, "tlsKeyFile", "/etc/lockvalidation/cert/key.pem", "File containing tls key")
	flag.Parse()
}

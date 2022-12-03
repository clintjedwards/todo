// Package api controls the bulk of the Todo API logic.
package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/clintjedwards/todo/internal/config"
	"github.com/clintjedwards/todo/internal/storage"
	proto "github.com/clintjedwards/todo/proto"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func ptr[T any](v T) *T {
	return &v
}

// API represents the main Todo service API. It is run using a GRPC/HTTP combined server.
type API struct {
	// Config represents the relative configuration for the Todo API. This is a combination of envvars and config values
	// gleaned at startup time.
	config *config.API

	// Storage represents the main backend storage implementation. Todo stores most of its critical state information
	// using this storage mechanism.
	db storage.DB

	// We opt out of forward compatibility with this embedded interface. This is required by GRPC.
	//
	// We don't embed the "proto.UnimplementedTodoServer" as there should never(I assume this will come back to bite me)
	// be an instance where we add proto methods without also updating the server to support those methods.
	// There is the added benefit that without it embedded we get compile time errors when a function isn't correctly
	// implemented. Saving us from weird "Unimplemented" RPC bugs.
	proto.UnsafeTodoServer
}

// NewAPI creates a new instance of the main Todo API service.
func NewAPI(config *config.API, storage storage.DB) (*API, error) {
	newAPI := &API{
		config: config,
		db:     storage,
	}

	return newAPI, nil
}

// StartAPIService starts the Todo API service and blocks until a SIGINT or SIGTERM is received.
func (api *API) StartAPIService() {
	grpcServer, err := api.createGRPCServer()
	if err != nil {
		log.Fatal().Err(err).Msg("could not create GRPC service")
	}

	tlsConfig, err := api.generateTLSConfig(api.config.Server.TLSCertPath, api.config.Server.TLSKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get proper TLS config")
	}

	httpServer := wrapGRPCServer(api.config, grpcServer)
	httpServer.TLSConfig = tlsConfig

	// Run our server in a goroutine and listen for signals that indicate graceful shutdown
	go func() {
		if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server exited abnormally")
		}
	}()
	log.Info().Str("url", api.config.Server.Host).Msg("started todo grpc/http service")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c

	// Doesn't block if no connections, otherwise will wait until the timeout deadline or connections to finish,
	// whichever comes first.
	ctx, cancel := context.WithTimeout(context.Background(), api.config.Server.ShutdownTimeout) // shutdown gracefully
	defer cancel()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not shutdown server in timeout specified")
		return
	}

	log.Info().Msg("grpc server exited gracefully")
}

// wrapGRPCServer returns a combined grpc/http (grpc-web compatible) service with all proper settings;
// Rather than going through the trouble of setting up a separate proxy and extra for the service in order to server http/grpc/grpc-web
// this keeps things simple by enabling the operator to deploy a single binary and serve them all from one endpoint.
// This reduces operational burden, configuration headache and overall just makes for a better time for both client and operator.
func wrapGRPCServer(config *config.API, grpcServer *grpc.Server) *http.Server {
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	router := mux.NewRouter()

	combinedHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Content-Type"), "application/grpc") || wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}
		router.ServeHTTP(resp, req)
	})

	var modifiedHandler http.Handler
	if config.DevMode {
		modifiedHandler = handlers.LoggingHandler(os.Stdout, combinedHandler)
	} else {
		modifiedHandler = combinedHandler
	}

	httpServer := http.Server{
		Addr:    config.Server.Host,
		Handler: modifiedHandler,
		// Timeouts set here unfortunately also apply to the backing GRPC server. Because GRPC might have long running calls
		// we have to set these to 0 or a very high number. This creates an issue where running the frontend in this configuration
		// could possibly open us up to DOS attacks where the client holds the request open for long periods of time. To mitigate
		// this we both implement timeouts for routes on both the GRPC side and the pure HTTP side.
		WriteTimeout: 0,
		ReadTimeout:  0,
	}

	return &httpServer
}

// createGRPCServer creates the todo grpc server with all the proper settings; TLS enabled.
func (api *API) createGRPCServer() (*grpc.Server, error) {
	tlsConfig, err := api.generateTLSConfig(api.config.Server.TLSCertPath, api.config.Server.TLSKeyPath)
	if err != nil {
		return nil, err
	}

	panicHandler := func(p interface{}) (err error) {
		log.Error().Err(err).Interface("panic", p).Bytes("stack", debug.Stack()).Msg("server has encountered a fatal error")
		return status.Errorf(codes.Unknown, "server has encountered a fatal error and could not process request")
	}

	grpcServer := grpc.NewServer(
		// recovery should always be first
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(panicHandler)),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(panicHandler)),
			),
		),

		// Handle TLS
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	reflection.Register(grpcServer)
	proto.RegisterTodoServer(grpcServer, api)

	return grpcServer, nil
}

// grpcDial establishes a connection with the request URL via GRPC.
func grpcDial(url string) (*grpc.ClientConn, error) {
	host, port, ok := strings.Cut(url, ":")
	if !ok {
		return nil, fmt.Errorf("could not parse url %q; format should be <host>:<port>", url)
	}

	var opt []grpc.DialOption
	var tlsConf *tls.Config

	// If we're testing in development bypass the cert checks.
	if host == "localhost" || host == "127.0.0.1" {
		tlsConf = &tls.Config{
			InsecureSkipVerify: true,
		}
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	}

	opt = append(opt, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(3), grpc_retry.WithBackoff(grpc_retry.BackoffExponential(time.Millisecond*100)))))

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), opt...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %w", err)
	}

	return conn, nil
}

// We use these functions to supply TLS for various services that require it. To make development easy
// we bake in general localhost certs for quick bootstrap. The server will not start with dev certs loaded
// unless explicitly told to do so with devmode=true.

//go:embed localhost.crt
var devtlscert []byte

//go:embed localhost.key
var devtlskey []byte

// generateTLSConfig returns TLS config object necessary for HTTPS loaded from files. If server is in devmode and
// no cert is provided it instead loads certificates from embedded files for ease of development.
func (api *API) generateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	var serverCert tls.Certificate
	var err error

	if api.config.DevMode && certPath == "" {
		serverCert, err = tls.X509KeyPair(devtlscert, devtlskey)
		if err != nil {
			return nil, err
		}
	} else {
		if certPath == "" || keyPath == "" {
			return nil, fmt.Errorf("TLS cert and key cannot be empty")
		}

		serverCert, err = tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return tlsConfig, nil
}

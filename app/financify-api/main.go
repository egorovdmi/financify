package main

import (
	"context"
	"crypto/rsa"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/egorovdmi/financify/app/financify-api/handlers"
	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/foundation/database"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "FINANCIFY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =============================================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Auth struct {
			KeyID          string `conf:"default:90a50c59-e095-4c36-b9a3-54f83a3832e2"`
			PrivateKeyFile string `conf:"default:./private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		DB struct {
			Host       string `conf:"default:database-service"`
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
		}
		Zipkin struct {
			ReporterURI string  `conf:"default:http://localhost:9411/api/v2/spans"`
			ServiceName string  `conf:"default:financify-api"`
			Probability float64 `conf:"default:0.05"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	if err := conf.Parse(os.Args[1:], "FINANCIFY", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("FINANCIFY", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}

			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("FINANCIFY", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}

			fmt.Println(version)
			return nil
		}

		return errors.Wrap(err, "parsing config")
	}

	expvar.NewString("build").Set(build)
	log.Printf("main: started: application initializing : version %q\n", build)
	defer log.Println("main: completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: config: \n%v\n", out)

	// =============================================================================================
	// Initialize authentication support

	log.Println("main: initializing authentication support")

	privatePEM, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}

	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case cfg.Auth.KeyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}

	auth, err := auth.New(cfg.Auth.Algorithm, lookup, auth.Keys{cfg.Auth.KeyID: privateKey})
	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}

	// =============================================================================================
	// Initialize database support
	log.Println("main: initializing database support")
	db, err := database.Open(database.Config{
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		DisableTLS: cfg.DB.DisableTLS,
	})

	if err != nil {
		return errors.Wrap(err, "connection to db")
	}

	defer func() {
		log.Printf("main: Database stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	log.Printf("main: initializing OT/Zipkin tracing support")

	traceProvider, err := startTracing(
		cfg.Zipkin.ServiceName,
		cfg.Zipkin.ReporterURI,
		cfg.Zipkin.Probability,
	)
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}
	defer traceProvider.Shutdown(context.Background())

	// =============================================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to default mux by importing the net/http/pprof packege.
	// /debug/vars - Added to default mux by importing the expvar packege.

	log.Println("main: initializing debugging support")

	go func() {
		log.Printf("main: debug listening %s", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: debug listener closed: %v", err)
		}
	}()

	// =============================================================================================
	// Start API Service

	log.Println("main: initializing API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log, auth, db),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main: %v: start shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}

// startTracing configure open telemetry to be used with zipkin.
func startTracing(serviceName string, reporterURI string, probability float64) (*trace.TracerProvider, error) {

	// WARNING: The current settings are using defaults which may not be
	// compatible with your project. Please review the documentation for
	// opentelemetry.

	exporter, err := zipkin.New(
		reporterURI,
		// zipkin.WithLogger(zap.NewStdLog(log)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				attribute.String("exporter", "zipkin"),
			),
		),
	)

	// I can only get this working properly using the singleton :(
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

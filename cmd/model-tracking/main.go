package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/metalmatze/signal/healthcheck"
	"github.com/metalmatze/signal/internalserver"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"

	v1alpha1 "github.com/connylabs/model-tracking/api/v1alpha1"
	"github.com/connylabs/model-tracking/store"
	"github.com/connylabs/model-tracking/version"
)

const (
	logLevelAll   = "all"
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelNone  = "none"

	logFmtJson = "json"
	logFmtFmt  = "fmt"
)

var (
	availableLogLevels = strings.Join([]string{
		logLevelAll,
		logLevelDebug,
		logLevelInfo,
		logLevelWarn,
		logLevelError,
		logLevelNone,
	}, ", ")

	availableLogFmts = strings.Join([]string{
		logFmtJson,
		logFmtFmt,
	}, ",")
)

// Main is the principal function for the binary, wrapped only by `main` for convenience.
func Main() error {
	postgresURL := flag.String("database", "", "Database connection string")
	listen := flag.String("listen", ":8080", "The address at which to listen.")
	listenInternal := flag.String("listen-internal", ":9090", "The address at which to listen for health and metrics.")
	healthCheckURL := flag.String("healthchecks-url", "http://localhost:8080", "The URL against which to run healthchecks.")
	logLevel := flag.String("log-level", logLevelInfo, fmt.Sprintf("Log level to use. Possible values: %s", availableLogLevels))
	logFmt := flag.String("log-fmt", logFmtFmt, fmt.Sprintf("Log format to use. Possible values: %s", availableLogFmts))
	help := flag.Bool("h", false, "Show usage")
	printVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	if *help {
		flag.Usage()
		return nil
	}

	if *printVersion {
		fmt.Println(version.Version)
		return nil
	}

	var logger log.Logger
	switch *logFmt {
	case logFmtJson:
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	case logFmtFmt:
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	default:
		return fmt.Errorf("log format %v unknown; possible values are: %s", *logFmt, availableLogFmts)
	}

	switch *logLevel {
	case logLevelAll:
		logger = level.NewFilter(logger, level.AllowAll())
	case logLevelDebug:
		logger = level.NewFilter(logger, level.AllowDebug())
	case logLevelInfo:
		logger = level.NewFilter(logger, level.AllowInfo())
	case logLevelWarn:
		logger = level.NewFilter(logger, level.AllowWarn())
	case logLevelError:
		logger = level.NewFilter(logger, level.AllowError())
	case logLevelNone:
		logger = level.NewFilter(logger, level.AllowNone())
	default:
		return fmt.Errorf("log level %v unknown; possible values are: %s", *logLevel, availableLogLevels)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	stdlog.SetOutput(log.NewStdlibAdapter(logger))

	if *postgresURL == "" {
		return errors.New("a value for --database must be specified")
	}

	db, err := sql.Open("pgx", *postgresURL)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	var g run.Group
	{
		l, err := net.Listen("tcp", *listen)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %v", *listen, err)
		}

		g.Add(func() error {
			level.Info(logger).Log("msg", "starting the model-tracking HTTP server", "addr", *listen, "version", version.Version)
			r := chi.NewRouter()
			v1alpha1.HandlerWithOptions(v1alpha1.NewServer(store.NewSQLStore(db), log.With(logger, "component", "http-server")), v1alpha1.ChiServerOptions{
				BaseRouter: r,
				BaseURL:    "/api/v1alpha1",
			})
			//v1alpha1.New(
			//log.With(logger, "component", "http-server"),
			//v1alpha1.WithRegisterer(prometheus.WrapRegistererWith(prometheus.Labels{"api": "v1"}, reg)),
			//v1alpha1.WithRouter(r),
			//)
			if err := http.Serve(l, r); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("error: server exited unexpectedly: %v", err)
			}
			return nil
		}, func(error) {
			l.Close()
		})
	}

	{
		// Run the internal HTTP server.
		healthchecks := healthcheck.NewMetricsHandler(healthcheck.NewHandler(), reg)
		// Checks if the server is up.
		healthchecks.AddLivenessCheck("http",
			healthcheck.HTTPCheckClient(
				http.DefaultClient,
				*healthCheckURL,
				http.MethodGet,
				http.StatusNotFound,
				time.Second,
			),
		)
		h := internalserver.NewHandler(
			internalserver.WithName("Internal - model-tracking"),
			internalserver.WithHealthchecks(healthchecks),
			internalserver.WithPrometheusRegistry(reg),
			internalserver.WithPProf(),
		)
		l, err := net.Listen("tcp", *listenInternal)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %v", *listenInternal, err)
		}

		g.Add(func() error {
			level.Info(logger).Log("msg", "starting the model-tracking internal HTTP server", "addr", *listenInternal)

			if err := http.Serve(l, h); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("error: internal server exited unexpectedly: %v", err)
			}
			return nil
		}, func(error) {
			l.Close()
		})
	}

	{
		// Exit gracefully on SIGINT and SIGTERM.
		term := make(chan os.Signal, 1)
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(func() error {
			for {
				select {
				case <-term:
					level.Info(logger).Log("msg", "caught interrupt; gracefully cleaning up; see you next time!")
					return nil
				case <-cancel:
					return nil
				}
			}
		}, func(error) {
			close(cancel)
		})
	}

	return g.Run()
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

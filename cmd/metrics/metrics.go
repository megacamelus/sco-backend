package metrics

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/sco1237896/sco-backend/pkg/metrics/publisher/stdout"

	gexpvar "github.com/gin-contrib/expvar"
	"github.com/gin-contrib/pprof"
	sloggin "github.com/samber/slog-gin"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/gin-gonic/gin"

	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/metrics/collector"
	"github.com/sco1237896/sco-backend/pkg/metrics/publisher"
	expvarsrv "github.com/sco1237896/sco-backend/pkg/metrics/publisher/expvar"
	prometheussrv "github.com/sco1237896/sco-backend/pkg/metrics/publisher/prometheus"
	"github.com/spf13/cobra"
)

const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 5 * time.Second
)

type srv struct {
	host            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

type prometheus struct {
	srv
	route string
}

type expvar struct {
	srv
	route string
}

type collect struct {
	from string
}

type publish struct {
	to       string
	interval time.Duration
}

type configs struct {
	development bool
	web         srv
	expvar      expvar
	prometheus  prometheus
	collect     collect
	publish     publish
}

var build = "develop"

func NewMetricsCmd() *cobra.Command {
	cfg := configs{
		development: false,
		web: srv{
			host:            "0.0.0.0:9003",
			readTimeout:     readTimeout,
			writeTimeout:    writeTimeout,
			idleTimeout:     idleTimeout,
			shutdownTimeout: shutdownTimeout,
		},
		expvar: expvar{
			route: "/metrics",
			srv: srv{
				host:            "0.0.0.0:9001",
				readTimeout:     readTimeout,
				writeTimeout:    writeTimeout,
				idleTimeout:     idleTimeout,
				shutdownTimeout: shutdownTimeout,
			},
		},
		prometheus: prometheus{
			route: "/metrics",
			srv: srv{
				host:            "0.0.0.0:9002",
				readTimeout:     readTimeout,
				writeTimeout:    writeTimeout,
				idleTimeout:     idleTimeout,
				shutdownTimeout: shutdownTimeout,
			},
		},
		collect: collect{
			from: "http://localhost:8083/debug/vars",
		},
		publish: publish{
			to:       "console",
			interval: 5 * time.Second,
		},
	}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "metrics",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(cfg.development)
			if !cfg.development {
				gin.SetMode(gin.ReleaseMode)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			g, ctx := errgroup.WithContext(cmd.Context())
			var log = logger.L.WithGroup("METRICS")

			// -------------------------------------------------------------------------
			// GOMAXPROCS
			_, err := maxprocs.Set(maxprocs.Logger(func(f string, a ...interface{}) { logger.L.Info(fmt.Sprintf(f, a)) }))
			if err != nil {
				logger.L.ErrorContext(ctx, "failed to set GOMAXPROCS from cgroups")
			}

			// -------------------------------------------------------------------------
			// App Starting

			log.InfoContext(ctx, "starting service", "version", build)
			defer log.InfoContext(ctx, "shutdown complete")

			log.InfoContext(ctx, "startup", "config", cfg)

			// -------------------------------------------------------------------------
			// Start Debug Service

			router := gin.Default()
			router.Use(gin.Recovery())
			router.Use(sloggin.New(logger.L.WithGroup("debug")))

			// register pprof middleware endpoints
			pprof.Register(router)

			// register expvar endpoints
			router.GET("/debug/vars", gexpvar.Handler())

			g.Go(func() error {
				logger.L.InfoContext(ctx, "startup", "status", "debug v1 router started", "host", cfg.web.host)

				srv := http.Server{
					Addr:         cfg.web.host,
					Handler:      router,
					ReadTimeout:  cfg.web.readTimeout,
					IdleTimeout:  cfg.web.idleTimeout,
					WriteTimeout: cfg.web.writeTimeout,
				}

				err = srv.ListenAndServe()
				if err != nil {
					return err
				}

				return nil
			})

			// -------------------------------------------------------------------------
			// Start Prometheus Service

			prom := prometheussrv.New(log, cfg.prometheus.host, cfg.prometheus.route, cfg.prometheus.readTimeout, cfg.prometheus.writeTimeout, cfg.prometheus.idleTimeout)
			defer prom.Stop(cfg.prometheus.shutdownTimeout)

			g.Go(func() error {
				log.InfoContext(ctx, "prometheus", "status", "API listening", "host", cfg.prometheus.host)
				return prom.Server.ListenAndServe()
			})

			// -------------------------------------------------------------------------
			// Start expvar Service

			exp := expvarsrv.New(log, cfg.expvar.host, cfg.expvar.route, cfg.expvar.readTimeout, cfg.expvar.writeTimeout, cfg.expvar.idleTimeout)
			defer exp.Stop(cfg.expvar.shutdownTimeout)

			g.Go(func() error {
				log.InfoContext(ctx, "expvar", "status", "API listening", "host", cfg.expvar.host)
				return exp.Server.ListenAndServe()
			})

			// -------------------------------------------------------------------------
			// Start collectors and publishers

			collector, err := collector.New(cfg.collect.from)
			if err != nil {
				return fmt.Errorf("starting collector: %w", err)
			}

			pstdout := stdout.NewStdout(log)

			publish, err := publisher.New(log, collector, cfg.publish.interval, prom.Publish, exp.Publish, pstdout.Publish)
			if err != nil {
				return fmt.Errorf("starting publisher: %w", err)
			}
			defer publish.Stop()

			// -------------------------------------------------------------------------
			// Shutdown

			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-shutdown:
				log.InfoContext(ctx, "shutdown", "status", "shutdown started")
				defer log.InfoContext(ctx, "shutdown", "status", "shutdown complete")

				return nil
			case <-ctx.Done():
				log.ErrorContext(ctx, "metrics", "could not start http server", "msg", err)
				return ctx.Err()
			}
		},
	}

	cmd.Flags().BoolVar(&cfg.development, "dev", cfg.development, "Turn on/off development mode")
	cmd.Flags().StringVar(&cfg.web.host, "debug-bind-address", cfg.web.host, "Main service debug address")
	cmd.Flags().StringVar(&cfg.expvar.host, "expvar-bind-address", cfg.expvar.host, "Expvar service bind address")
	cmd.Flags().StringVar(&cfg.prometheus.host, "prometheus-bind-address", cfg.prometheus.host, "Prometheus service bind address")
	cmd.Flags().StringVar(&cfg.collect.from, "collect", cfg.collect.from, "Main service address used to collect metrics from")

	return cmd
}

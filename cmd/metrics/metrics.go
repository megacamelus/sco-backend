package metrics

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/expvar"
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

type Web struct {
	DebugHost       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Expvar struct {
	Host            string
	Route           string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Prometheus struct {
	Host            string
	Route           string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Collect struct {
	From string
}

type Publish struct {
	To       string
	Interval time.Duration
}

type configs struct {
	Development bool
	Web
	Expvar
	Prometheus
	Collect
	Publish
}

var build = "develop"

func NewMetricsCmd() *cobra.Command {
	cfg := configs{
		Development: false,
		Web: Web{
			DebugHost:       "0.0.0.0:9003",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Expvar: Expvar{
			Host:            "0.0.0.0:9001",
			Route:           "/metrics",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Prometheus: Prometheus{
			Host:            "0.0.0.0:9002",
			Route:           "/metrics",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 5 + time.Second,
		},
		Collect: Collect{
			From: "http://localhost:8083/debug/vars",
		},
		Publish: Publish{
			To:       "console",
			Interval: 5 * time.Second,
		},
	}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "metrics",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logger.Init(cfg.Development)
			if !cfg.Development {
				gin.SetMode(gin.ReleaseMode)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
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
			router.GET("/debug/vars", expvar.Handler())

			go func() {
				logger.L.InfoContext(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

				srv := http.Server{
					Addr:         cfg.Web.DebugHost,
					Handler:      router,
					ReadTimeout:  cfg.Web.ReadTimeout,
					IdleTimeout:  cfg.Web.IdleTimeout,
					WriteTimeout: cfg.Web.WriteTimeout,
				}

				err = srv.ListenAndServe()
				if err != nil {
					logger.L.ErrorContext(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
				}
			}()

			// -------------------------------------------------------------------------
			// Start Prometheus Service

			prom := prometheussrv.New(log, cfg.Prometheus.Host, cfg.Prometheus.Route, cfg.Prometheus.ReadTimeout, cfg.Prometheus.WriteTimeout, cfg.Prometheus.IdleTimeout)
			defer prom.Stop(cfg.Prometheus.ShutdownTimeout)

			// -------------------------------------------------------------------------
			// Start expvar Service

			exp := expvarsrv.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout, cfg.Expvar.IdleTimeout)
			defer exp.Stop(cfg.Expvar.ShutdownTimeout)

			// -------------------------------------------------------------------------
			// Start collectors and publishers

			collector, err := collector.New(cfg.Collect.From)
			if err != nil {
				return fmt.Errorf("starting collector: %w", err)
			}

			stdout := publisher.NewStdout(log)

			publish, err := publisher.New(log, collector, cfg.Publish.Interval, prom.Publish, exp.Publish, stdout.Publish)
			if err != nil {
				return fmt.Errorf("starting publisher: %w", err)
			}
			defer publish.Stop()

			// -------------------------------------------------------------------------
			// Shutdown

			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
			<-shutdown

			log.InfoContext(ctx, "shutdown", "status", "shutdown started")
			defer log.InfoContext(ctx, "shutdown", "status", "shutdown complete")

			return nil
		},
	}

	cmd.Flags().BoolVar(&cfg.Development, "dev", cfg.Development, "Turn on/off development mode")
	cmd.Flags().StringVar(&cfg.Web.DebugHost, "debug-bind-address", cfg.Web.DebugHost, "Main service debug address")
	cmd.Flags().StringVar(&cfg.Expvar.Host, "expvar-bind-address", cfg.Expvar.Host, "Expvar service bind address")
	cmd.Flags().StringVar(&cfg.Prometheus.Host, "prometheus-bind-address", cfg.Prometheus.Host, "Prometheus service bind address")
	cmd.Flags().StringVar(&cfg.Collect.From, "collect", cfg.Collect.From, "Main service address used to collect metrics from")

	return cmd
}

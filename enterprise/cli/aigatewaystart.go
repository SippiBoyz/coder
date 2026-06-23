//go:build !slim

package cli

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/xerrors"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/sloggers/sloghuman"
	"github.com/coder/coder/v2/aibridge"
	agpl "github.com/coder/coder/v2/cli"
	"github.com/coder/coder/v2/coderd/aibridged"
	"github.com/coder/coder/v2/coderd/aibridged/proto"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/retry"
	"github.com/coder/serpent"
)

// aiGatewayStart runs the AI Gateway as a standalone process.
// It connects to coderd over DRPC via /api/v2/aibridge/serve for
// authentication, recording, and MCP configuration, and listens on its
// own HTTP address for incoming LLM client traffic. Providers are built
// from the deployment configuration; the standalone process does not read
// the database directly.
func (r *RootCmd) aiGatewayStart() *serpent.Command {
	var (
		key         string
		httpAddress string
		tlsCertFile string
		tlsKeyFile  string
		verbose     bool
	)

	// Reuse the shared AI Gateway deployment options (CODER_AI_GATEWAY_*)
	// so standalone mode is configured exactly like embedded mode. The
	// option Values point into vals, which is captured by the handler.
	vals := new(codersdk.DeploymentValues)
	var aiGatewayOpts serpent.OptionSet
	for _, opt := range vals.Options() {
		if opt.Group != nil && opt.Group.Name == "AI Gateway" {
			aiGatewayOpts = append(aiGatewayOpts, opt)
		}
	}

	cmd := &serpent.Command{
		Use:   "start",
		Short: "Run a standalone AI Gateway server",
		Long: "The standalone AI Gateway connects to a Coder deployment over DRPC to " +
			"authenticate users, record interceptions, and configure MCP, while serving " +
			"LLM client traffic on its own HTTP listener.\n\n" +
			"The deployment address is taken from the global --url flag (CODER_URL) and " +
			"is required. The gateway authenticates with the key from --key " +
			"(CODER_AI_GATEWAY_KEY). Provider and other AI Gateway settings use the same " +
			"CODER_AI_GATEWAY_* options as embedded mode.",
		Handler: func(inv *serpent.Invocation) error {
			ctx, cancel := context.WithCancel(inv.Context())
			defer cancel()

			if key == "" {
				return xerrors.New("an AI Gateway key is required; set --key or CODER_AI_GATEWAY_KEY")
			}
			// TLS is opt-in and requires both files; setting only one is
			// an error. Default is plain HTTP.
			if (tlsCertFile == "") != (tlsKeyFile == "") {
				return xerrors.New("--tls-cert-file and --tls-key-file must be provided together")
			}

			client, err := r.InitClient(inv)
			if err != nil {
				return err
			}

			notifyCtx, notifyStop := inv.SignalNotifyContext(ctx, agpl.StopSignals...)
			defer notifyStop()

			logger := slog.Make(sloghuman.Sink(inv.Stderr))
			if verbose {
				logger = logger.Leveled(slog.LevelDebug)
			}

			// Metrics and tracing are not yet exposed by standalone mode
			// (future work), but the pool requires a metrics object and a
			// tracer, so wire up no-op sinks.
			metrics := aibridge.NewMetrics(prometheus.NewRegistry())
			tracer := trace.NewNoopTracerProvider().Tracer("aibridged")

			// The standalone gateway has no provider env vars and no database
			// access. It starts with an empty pool, connects to coderd over
			// DRPC, then fetches the provider set via GetAIProviders and builds
			// the pool from it.
			pool, err := aibridged.NewCachedBridgePool(aibridged.DefaultPoolOptions, nil, logger.Named("pool"), metrics, tracer)
			if err != nil {
				return xerrors.Errorf("create request pool: %w", err)
			}

			dialer := aibridged.NewWebsocketDialer(client, key)
			srv, err := aibridged.New(ctx, pool, dialer, logger.Named("aibridged"), tracer)
			if err != nil {
				return xerrors.Errorf("start aibridge daemon: %w", err)
			}
			defer srv.Close()

			// Fetch the initial provider set from coderd, retrying until
			// success. srv.Client() blocks until the daemon connects; an empty
			// provider list is a valid result and ends the loop. The standalone
			// gateway has no refresh trigger until AIGOV-465, so this runs once
			// on startup.
			if err := initStandaloneProviders(ctx, srv, pool, vals.AI.BridgeConfig, logger.Named("aibridge.providers"), metrics); err != nil {
				return xerrors.Errorf("initialize ai providers: %w", err)
			}

			// The standalone listener is dedicated to Gateway traffic, so
			// the daemon is served at the root. The /api/v2/aibridge alias
			// keeps parity with the embedded route, so a Gateway proxy
			// pointed here with the embedded path still works.
			mux := http.NewServeMux()
			mux.Handle("/api/v2/aibridge/", http.StripPrefix("/api/v2/aibridge", srv))
			mux.Handle("/", srv)

			listener, err := net.Listen("tcp", httpAddress)
			if err != nil {
				return xerrors.Errorf("listen on %q: %w", httpAddress, err)
			}
			defer listener.Close()

			httpServer := &http.Server{
				Handler:           mux,
				ReadHeaderTimeout: time.Minute,
			}

			serveErr := make(chan error, 1)
			go func() {
				if tlsCertFile != "" {
					serveErr <- httpServer.ServeTLS(listener, tlsCertFile, tlsKeyFile)
				} else {
					serveErr <- httpServer.Serve(listener)
				}
			}()

			logger.Info(ctx, "standalone AI Gateway listening",
				slog.F("address", listener.Addr().String()),
				slog.F("tls", tlsCertFile != ""),
			)

			select {
			case <-notifyCtx.Done():
				logger.Info(ctx, "shutting down standalone AI Gateway")
			case err := <-serveErr:
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return xerrors.Errorf("serve: %w", err)
				}
			}

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				return xerrors.Errorf("shutdown http server: %w", err)
			}
			return nil
		},
	}

	cmd.Options = serpent.OptionSet{
		{
			Flag:        "key",
			Env:         "CODER_AI_GATEWAY_KEY",
			Description: "The AI Gateway key used to authenticate to coderd.",
			Value:       serpent.StringOf(&key),
		},
		{
			Flag:        "http-address",
			Env:         "CODER_AI_GATEWAY_HTTP_ADDRESS",
			Description: "The bind address to serve incoming AI Gateway client traffic.",
			Default:     "127.0.0.1:4001",
			Value:       serpent.StringOf(&httpAddress),
		},
		{
			Flag:        "tls-cert-file",
			Env:         "CODER_AI_GATEWAY_TLS_CERT_FILE",
			Description: "Path to a PEM-encoded TLS certificate. Enables TLS termination when set together with --tls-key-file.",
			Value:       serpent.StringOf(&tlsCertFile),
		},
		{
			Flag:        "tls-key-file",
			Env:         "CODER_AI_GATEWAY_TLS_KEY_FILE",
			Description: "Path to a PEM-encoded TLS private key. Enables TLS termination when set together with --tls-cert-file.",
			Value:       serpent.StringOf(&tlsKeyFile),
		},
		{
			Flag:        "verbose",
			Env:         "CODER_AI_GATEWAY_VERBOSE",
			Description: "Output debug-level logs.",
			Value:       serpent.BoolOf(&verbose),
			Default:     "false",
		},
	}
	cmd.Options = append(cmd.Options, aiGatewayOpts...)

	return cmd
}

// initStandaloneProviders fetches the AI provider set from coderd over DRPC
// and populates the pool, retrying until it succeeds or ctx is canceled.
// srv.Client() blocks until the daemon connects to coderd; GetAIProviders may
// still fail transiently (e.g. mid-seed contention or a dropped connection),
// so the whole fetch is retried with backoff. A successful empty list is a
// valid result and ends the loop.
//
// The standalone gateway has no provider-change refresh trigger until
// AIGOV-465, so this runs once on startup; provider add/enable will not
// propagate to a running standalone gateway.
func initStandaloneProviders(
	ctx context.Context,
	srv *aibridged.Server,
	pool *aibridged.CachedBridgePool,
	cfg codersdk.AIBridgeConfig,
	logger slog.Logger,
	metrics *aibridge.Metrics,
) error {
	for r := retry.New(50*time.Millisecond, 10*time.Second); r.Wait(ctx); {
		client, err := srv.Client()
		if err != nil {
			// Client() only fails when the daemon is shutting down.
			return xerrors.Errorf("get aibridge client: %w", err)
		}
		resp, err := client.GetAIProviders(ctx, &proto.GetAIProvidersRequest{})
		if err != nil {
			logger.Warn(ctx, "fetch ai providers, will retry", slog.Error(err))
			continue
		}
		providers, _ := agpl.BuildProvidersFromProto(ctx, resp.GetProviders(), cfg, logger, metrics)
		pool.ReplaceProviders(providers)
		logger.Info(ctx, "loaded ai providers from coderd", slog.F("count", len(providers)))
		return nil
	}
	return ctx.Err()
}

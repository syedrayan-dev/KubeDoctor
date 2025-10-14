/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	diagnosisv1alpha1 "github.com/yth01/apollo/api/v1alpha1"
	"github.com/yth01/apollo/internal/api"
	"github.com/yth01/apollo/internal/controller"
	"github.com/yth01/apollo/internal/utils"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

// apiServerRunnable wraps our API server to implement manager.Runnable
type apiServerRunnable struct {
	server *api.Server
	port   int
}

// Start implements manager.Runnable
func (a *apiServerRunnable) Start(ctx context.Context) error {
	return a.server.Start(ctx, a.port)
}

// Config holds all configuration for the Apollo operator
type Config struct {
	MetricsAddr          string
	MetricsCertPath      string
	MetricsCertName      string
	MetricsCertKey       string
	WebhookCertPath      string
	WebhookCertName      string
	WebhookCertKey       string
	EnableLeaderElection bool
	ProbeAddr            string
	SecureMetrics        bool
	EnableHTTP2          bool
	APIPort              int
}

// parseFlags parses command line flags and returns configuration
func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.MetricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&config.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&config.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&config.SecureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&config.WebhookCertPath, "webhook-cert-path", "",
		"The directory that contains the webhook certificate.")
	flag.StringVar(&config.WebhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&config.WebhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.StringVar(&config.MetricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&config.MetricsCertName, "metrics-cert-name", "tls.crt",
		"The name of the metrics server certificate file.")
	flag.StringVar(&config.MetricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&config.EnableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.IntVar(&config.APIPort, "api-port", 8080, "The port for the API server")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	return config
}

// setupTLSConfig sets up TLS configuration based on HTTP/2 settings
func setupTLSConfig(enableHTTP2 bool) []func(*tls.Config) {
	var tlsOpts []func(*tls.Config)

	if !enableHTTP2 {
		disableHTTP2 := func(c *tls.Config) {
			setupLog.Info("disabling http/2")
			c.NextProtos = []string{"http/1.1"}
		}
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	return tlsOpts
}

// setupWebhookServer sets up the webhook server with optional certificate watching
func setupWebhookServer(config *Config, tlsOpts []func(*tls.Config)) (webhook.Server, error) {
	webhookTLSOpts := tlsOpts

	if len(config.WebhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", config.WebhookCertPath,
			"webhook-cert-name", config.WebhookCertName,
			"webhook-cert-key", config.WebhookCertKey)

		webhookCertWatcher, err := certwatcher.New(
			filepath.Join(config.WebhookCertPath, config.WebhookCertName),
			filepath.Join(config.WebhookCertPath, config.WebhookCertKey),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize webhook certificate watcher: %w", err)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(c *tls.Config) {
			c.GetCertificate = webhookCertWatcher.GetCertificate
		})
	}

	return webhook.NewServer(webhook.Options{TLSOpts: webhookTLSOpts}), nil
}

// setupMetricsServer sets up the metrics server with optional certificate watching
func setupMetricsServer(config *Config, tlsOpts []func(*tls.Config)) (metricsserver.Options, error) {
	metricsServerOptions := metricsserver.Options{
		BindAddress:   config.MetricsAddr,
		SecureServing: config.SecureMetrics,
		TLSOpts:       tlsOpts,
	}

	if len(config.MetricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", config.MetricsCertPath,
			"metrics-cert-name", config.MetricsCertName,
			"metrics-cert-key", config.MetricsCertKey)

		metricsCertWatcher, err := certwatcher.New(
			filepath.Join(config.MetricsCertPath, config.MetricsCertName),
			filepath.Join(config.MetricsCertPath, config.MetricsCertKey),
		)
		if err != nil {
			return metricsServerOptions, fmt.Errorf("failed to initialize metrics certificate watcher: %w", err)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(c *tls.Config) {
			c.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}

	return metricsServerOptions, nil
}

// setupControllers sets up all controllers with the manager
func setupControllers(mgr ctrl.Manager) error {
	if err := (&controller.PodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create Pod controller: %w", err)
	}

	if err := (&controller.DiagnosisRequestReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create DiagnosisRequest controller: %w", err)
	}

	if err := (&controller.DiagnosisPolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create DiagnosisPolicy controller: %w", err)
	}

	if err := (&controller.DiagnosisReportReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create DiagnosisReport controller: %w", err)
	}

	return nil
}

// setupHealthProbes sets up health check endpoints
func setupHealthProbes(mgr ctrl.Manager) error {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	return nil
}

// runApollo is the main application logic, separated from main() for better testing
func runApollo() error {
	config := parseFlags()

	// Initialize scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(diagnosisv1alpha1.AddToScheme(scheme))

	// Setup TLS configuration
	tlsOpts := setupTLSConfig(config.EnableHTTP2)

	// Setup webhook server
	webhookServer, err := setupWebhookServer(config, tlsOpts)
	if err != nil {
		return err
	}

	// Setup metrics server
	metricsServerOptions, err := setupMetricsServer(config, tlsOpts)
	if err != nil {
		return err
	}

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsServerOptions,
		WebhookServer:                 webhookServer,
		HealthProbeBindAddress:        config.ProbeAddr,
		LeaderElection:                config.EnableLeaderElection,
		LeaderElectionID:              "cff50994.apollo.dev",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// Setup controllers
	if err := setupControllers(mgr); err != nil {
		return err
	}

	// Setup health probes
	if err := setupHealthProbes(mgr); err != nil {
		return err
	}

	// Setup API server
	apiServer := api.NewServer(mgr.GetClient())
	apiRunnable := &apiServerRunnable{
		server: apiServer,
		port:   config.APIPort,
	}

	if err := mgr.Add(apiRunnable); err != nil {
		return fmt.Errorf("unable to add API server to manager: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

func main() {
	if err := runApollo(); err != nil {
		utils.HandleFatalError(setupLog, err, "Apollo operator failed to start")
	}
}

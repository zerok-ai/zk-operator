package main

import (

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zkConfig "github.com/zerok-ai/zk-utils-go/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"time"

	"github.com/zerok-ai/zk-operator/internal"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/controllers"
	"github.com/zerok-ai/zk-operator/internal/handler"
	"github.com/zerok-ai/zk-operator/internal/server"

	"github.com/kataras/iris/v12"

	"github.com/zerok-ai/zk-operator/internal/config"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	LOG_TAG  = "Main"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var d time.Duration = 15 * time.Minute

	setupLog.Info("Starting Operator.")
	zkCRDProbeHandler, err := initOperator()
	if err != nil {
		message := "Failed to initialize operator with error " + err.Error()
		setupLog.Info(message)
		return
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		Namespace:              "",
		SyncPeriod:             &d,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		panic("unable to start manager")
	}

	if err = (&controllers.ZerokProbeReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		ZkCRDProbeHandler: zkCRDProbeHandler,
		Recorder:          mgr.GetEventRecorderFor("zerok-probe-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ZerokProbe")
		panic("unable to create controller")
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		panic("unable to set up health check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		panic("unable to set up ready check")
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		panic("problem running manager")
	}
}

func initOperator() (*handler.ZkCRDProbeHandler, error) {

	var cfg config.ZkOperatorConfig
	if err := zkConfig.ProcessArgs[config.ZkOperatorConfig](&cfg); err != nil {
		panic(err)
	}

	zklogger.Init(cfg.LogsConfig)

	zklogger.Debug(LOG_TAG, "Successfully read configs.")

	zkModules := make([]internal.ZkOperatorModule, 0)

	crdProbeHandler := handler.ZkCRDProbeHandler{}
	err := crdProbeHandler.Init(cfg)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating scenarioHandler ", err)
		return nil, err
	}

	//Adding crdProbeHandler to zkModules
	zkModules = append(zkModules, &crdProbeHandler)

	irisConfig := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              cfg.LogsConfig.Level,
	})

	app := newApp()

	// start http server
	go server.StartHttpServer(app, irisConfig, cfg, zkModules)

	return &crdProbeHandler, nil
}

func newApp() *iris.Application {
	app := iris.Default()

	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods", "POST")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
	app.UseRouter(crs)
	app.AllowMethods(iris.MethodOptions)
	ph, _ := getProbeHandler(config.AppConfig{})
	probe.Initialize(app.Party("/v1"), ph)

	//scraping metrics for prometheus
	app.Get("/metrics", iris.FromStd(promhttp.Handler()))

	//scraping metrics for prometheus
	app.Get("/metrics", iris.FromStd(promhttp.Handler()))

	return app
}

func getProbeHandler(cfg config.AppConfig) (probeHandler.ProbeHandler, error) {
	serviceStore := store.GetServiceStore(cfg.Redis)
	probeSvc := probeService.NewProbeService(serviceStore)
	return probeHandler.NewProbeHandler(probeSvc), nil
}

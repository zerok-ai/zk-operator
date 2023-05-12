/*
Copyright 2022.

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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"flag"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"fmt"
	"log"
	"os"

	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/controllers"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/pkg/cert"
	"github.com/zerok-ai/zk-operator/pkg/server"
	"github.com/zerok-ai/zk-operator/pkg/storage"
	"github.com/zerok-ai/zk-operator/pkg/sync"
	"github.com/zerok-ai/zk-operator/pkg/utils"

	"github.com/kataras/iris/v12"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var d time.Duration = 15 * time.Minute

	setupLog.Info("Starting injector.")
	initInjector()

	//Starting exception server
	go server.StartExceptionServer()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "96feec81.zerok.ai",
		Namespace:              "",
		SyncPeriod:             &d,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ZerokopReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Zerokop")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// TODO:
// Add zklogger in the project.
// Unit testing.
func initInjector() {

	setupLog.Info("Running injector main.")

	configPath := env.GetString("CONFIG_FILE", "")
	if configPath == "" {
		fmt.Println("Config yaml path not found.")
		return
	}

	var cfg config.ZkInjectorConfig

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Println(err)
	}

	runtimeMap := &storage.ImageRuntimeHandler{}
	runtimeMap.Init(cfg)
	go sync.UpdateOrchestration(runtimeMap, cfg)

	app := newApp()

	irisConfig := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              "debug",
	})

	go server.StartZkCloudServer(newApp(), cfg, irisConfig)

	// initialize certificates
	caPEM, cert, key, err := cert.InitializeKeysAndCertificates(cfg.Webhook)
	if err != nil {
		msg := fmt.Sprintf("Failed to create keys and certificates for webhook %v. Stopping initialization of the pod.\n", err)
		fmt.Println(msg)
		return
	}

	// start mutating webhook
	err = utils.CreateOrUpdateMutatingWebhookConfiguration(caPEM, cfg.Webhook)
	if err != nil {
		msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v. Stopping initialization of the pod.\n", err)
		fmt.Println(msg)
		return
	}

	// start webhook server
	go server.StartWebHookServer(app, cfg, cert, key, runtimeMap, irisConfig)

	setupLog.Info("End of running injector main.")
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

	return app
}

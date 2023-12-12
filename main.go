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
	"fmt"
	"os"
	"time"

	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/auth"

	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	handler "github.com/zerok-ai/zk-operator/internal/handler"
	server "github.com/zerok-ai/zk-operator/internal/server"
	"github.com/zerok-ai/zk-operator/internal/storage"

	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/controllers"

	"github.com/ilyakaznacheev/cleanenv"

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
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var d time.Duration = 15 * time.Minute
	setupLog.Info("Starting Operator.")
	//_, err := initOperator()
	//if err != nil {
	//	message := "Failed to initialize operator with error " + err.Error()
	//	setupLog.Info(message)
	//	return
	//}

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
		panic("unable to start manager")
	}

	//if err = (&controllers.ZerokopReconciler{
	//	Client: mgr.GetClient(),
	//	Scheme: mgr.GetScheme(),
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "Zerokop")
	//	panic("unable to create controller")
	//}

	//initializing zkCRDProbeHandler
	zkCRDProbeHandler, err := getCRDProbeHandler()
	if err = (&controllers.ZerokProbeReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		ZkCRDProbeHandler: zkCRDProbeHandler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ZerokProbe")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	//if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
	//	setupLog.Error(err, "unable to set up health check")
	//	panic("unable to set up health check")
	//}
	//if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
	//	setupLog.Error(err, "unable to set up ready check")
	//	panic("unable to set up ready check")
	//}
	//
	//setupLog.Info("starting manager")
	//if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
	//	setupLog.Error(err, "problem running manager")
	//	panic("problem running manager")
	//}
}

// TODO: Unit testing.
// TODO: Split the initOperator method to smaller sub-methods per handler. (e.g. initWebhook, initScenarioHandler, etc.)
func initOperator() ([]internal.ZkOperatorModule, error) {

	configPath := env.GetString("CONFIG_FILE", "")
	if configPath == "" {
		zklogger.Error(LOG_TAG, "Config yaml path not found.")
		return nil, fmt.Errorf("config yaml path not found")
	}

	var zkConfig config.ZkOperatorConfig

	if err := cleanenv.ReadConfig(configPath, &zkConfig); err != nil {
		zklogger.Error(LOG_TAG, "Error while reading config ", err)
		return nil, err
	}

	zklogger.Init(zkConfig.LogsConfig)

	zklogger.Debug(LOG_TAG, "Successfully read configs.")

	zkModules := make([]internal.ZkOperatorModule, 0)

	irisConfig := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              zkConfig.LogsConfig.Level,
	})

	zklogger.Debug(LOG_TAG, "Creating operation login handler.")

	//Creating operator login module
	opLogin := auth.CreateOperatorLogin(zkConfig)

	zklogger.Debug(LOG_TAG, "Creating scenario handler.")

	//Module for syncing scenarios
	scenarioHandler := handler.ScenarioHandler{}
	err := scenarioHandler.Init(opLogin, zkConfig)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating scenarioHandler ", err)
		return nil, err
	}
	zkModules = append(zkModules, &scenarioHandler)

	zklogger.Debug(LOG_TAG, "Creating obfuscation handler.")

	// Module for managing obfuscation rules
	obfuscationHandler := handler.ObfuscationHandler{}
	err = obfuscationHandler.Init(opLogin, zkConfig)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating obfuscationHandler ", err)
		return nil, err
	}

	zkModules = append(zkModules, &obfuscationHandler)

	zklogger.Debug(LOG_TAG, "Creating integrations handler.")

	//Module for syncing integrations
	integrationHandler := handler.IntegrationsHandler{}
	err = integrationHandler.Init(opLogin, zkConfig)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating integrationHandler ", err)
		return nil, err
	}
	zkModules = append(zkModules, &integrationHandler)

	zklogger.Debug(LOG_TAG, "Creating service config handler.")

	//Module for syncing integrations
	serviceConfigHandler := handler.ServiceConfigHandler{}
	err = serviceConfigHandler.Init(opLogin, zkConfig)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating serviceConfigHandler ", err)
		return nil, err
	}
	zkModules = append(zkModules, &serviceConfigHandler)

	clusterContextHandler := handler.ClusterContextHandler{OpLogin: opLogin, ZkConfig: &zkConfig}
	zkModules = append(zkModules, &clusterContextHandler)

	// Module for syncing Executor Attributes Map into Redis.
	executorAttributesHandler := handler.ExecutorAttributesHandler{}
	executorAttributesStore := storage.GetExecutorAttributesRedisStore(zkConfig)

	executorAttributesHandler.Init(executorAttributesStore, opLogin, zkConfig)
	zkModules = append(zkModules, &executorAttributesHandler)

	go opLogin.RegisterZkModules(zkModules)

	//Staring syncing scenarios from zk cloud.
	go scenarioHandler.StartPeriodicSync()

	//Staring syncing obfuscations from zk cloud.
	go obfuscationHandler.StartPeriodicSync()

	//Staring syncing integrations from zk cloud.
	go integrationHandler.StartPeriodicSync()

	//Staring syncing configurations from zk cloud.
	go serviceConfigHandler.StartPeriodicSync()

	//Staring syncing Exxecutor Attributes from zk cloud.
	go executorAttributesHandler.StartPeriodicSync()

	zklogger.Debug(LOG_TAG, "Starting webhook server.")

	app := newApp()

	// start http server
	go server.StartHttpServer(app, irisConfig, zkConfig, &clusterContextHandler, &serviceConfigHandler, zkModules)

	return zkModules, nil
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

func getCRDProbeHandler() (*handler.ZkCRDProbeHandler, error) {
	configPath := env.GetString("CONFIG_FILE", "")
	if configPath == "" {
		zklogger.Error(LOG_TAG, "Config yaml path not found.")
		return nil, fmt.Errorf("config yaml path not found")
	}

	var zkConfig config.ZkOperatorConfig

	if err := cleanenv.ReadConfig(configPath, &zkConfig); err != nil {
		zklogger.Error(LOG_TAG, "Error while reading config ", err)
		return nil, err
	}

	crdProbeHandler := handler.ZkCRDProbeHandler{}
	err := crdProbeHandler.Init(zkConfig)
	if err != nil {
		zklogger.Error(LOG_TAG, "Error while creating scenarioHandler ", err)
		return nil, err
	}
	return &crdProbeHandler, nil
}

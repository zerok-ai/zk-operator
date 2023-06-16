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
	"github.com/zerok-ai/zk-operator/internal"
	"github.com/zerok-ai/zk-operator/internal/auth"
	"github.com/zerok-ai/zk-operator/internal/common"
	"github.com/zerok-ai/zk-operator/internal/webhook"
	"github.com/zerok-ai/zk-utils-go/scenario/model"
	"time"

	handler "github.com/zerok-ai/zk-operator/internal/handler"
	server "github.com/zerok-ai/zk-operator/internal/server"
	"github.com/zerok-ai/zk-operator/internal/storage"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	zkredis "github.com/zerok-ai/zk-utils-go/storage/redis"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/env"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/controllers"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/kataras/iris/v12"
	"github.com/zerok-ai/zk-operator/internal/config"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var LOG_TAG = "Main"

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

	setupLog.Info("Starting Operator.")
	initOperator()

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

	if err = (&controllers.ZerokopReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Zerokop")
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

// TODO:
// Unit testing.
func initOperator() {

	configPath := env.GetString("CONFIG_FILE", "")
	if configPath == "" {
		fmt.Println("Config yaml path not found.")
		return
	}

	var zkConfig config.ZkOperatorConfig

	if err := cleanenv.ReadConfig(configPath, &zkConfig); err != nil {
		fmt.Println("Error while reading config ", err)
		return
	}

	zklogger.Init(zkConfig.LogsConfig)

	zkModules := make([]internal.ZkOperatorModule, 0)

	// initialize certificates
	caPEM, cert, key, err := webhook.InitializeKeysAndCertificates(zkConfig.Webhook)
	if err != nil {
		msg := fmt.Sprintf("Failed to create keys and certificates for webhook %v. Stopping initialization of the pod.\n", err)
		zklogger.Error(LOG_TAG, msg)
		return
	}

	app := newApp()

	irisConfig := iris.WithConfiguration(iris.Configuration{
		DisablePathCorrection: true,
		LogLevel:              zkConfig.LogsConfig.Level,
	})

	// creating mutating webhook
	webhookHandler := webhook.WebhookHandler{}
	webhookHandler.Init(caPEM, zkConfig.Webhook)
	zkModules = append(zkModules, &webhookHandler)

	//creating in-memory <image,runtime> map handler.
	imageRuntimeCache := &storage.ImageRuntimeCache{}
	imageRuntimeCache.Init(zkConfig)
	zkModules = append(zkModules, imageRuntimeCache)

	//Creating operator login module
	opLogin := auth.CreateOperatorLogin(zkConfig.OperatorLogin)

	//Module for syncing rules
	scenarioHandler := handler.ScenarioHandler{}
	versionedStore, err := zkredis.GetVersionedStore[model.Scenario](&zkConfig.Redis, common.RedisVersionDbName, true, model.Scenario{})
	if err != nil {
		//logger.ZkLogger.Err(LOG_TAG, "Error while creating versionedStore ", err.Error())
		return
	}
	scenarioHandler.Init(versionedStore, opLogin, zkConfig)
	zkModules = append(zkModules, &scenarioHandler)

	clusterConfigHandler := handler.ClusterConfigHandler{OpLogin: opLogin}
	zkModules = append(zkModules, &clusterConfigHandler)

	opLogin.RegisterZkModules(zkModules)

	//Starting syncing of image,runtime data from redis
	go imageRuntimeCache.StartPeriodicSync()

	//Staring syncing scenarios from zk cloud.
	go scenarioHandler.StartPeriodicSync()

	// start webhook server
	go server.StartWebHookServer(app, zkConfig, cert, key, imageRuntimeCache, irisConfig)

	// start http server
	go server.StartHttpServer(app, irisConfig, zkConfig.Http, &clusterConfigHandler)

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

package utils

import (
	"github.com/zerok-ai/zk-operator/internal/config"
	logger "github.com/zerok-ai/zk-utils-go/logs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"time"
)

var FINALIZER_LOG_TAG = "Finalizer"
var targetNamespace = "zk-client"
var targetFinalizer = "operator/cleanup-pods"

func ListenToNamespaceDeletion(config *config.ZkOperatorConfig) {
	clientSet, err := GetK8sClient()
	if err != nil {
		logger.Debug(FINALIZER_LOG_TAG, "Failed to create clientSet in listenToNamespaceDeletion")
		return
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Second*30, informers.WithNamespace(targetNamespace))

	namespaceInformer := informerFactory.Core().V1().Namespaces().Informer()

	_, err = namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			namespace, ok := obj.(*corev1.Namespace)
			if !ok {
				logger.Error(FINALIZER_LOG_TAG, "Failed to cast object to Namespace")
				return
			}
			logger.Debug(FINALIZER_LOG_TAG, "Namespace %s is deleted.\n", namespace.Name)
			err = cleanUpOrchestratedPods(config, namespace)
			if err != nil {
				logger.Error(FINALIZER_LOG_TAG, "Failed to clean up orchestrated pods")
				return
			}
		},
	})
	if err != nil {
		logger.Error(FINALIZER_LOG_TAG, "Failed to register event handler for namespace informer")
		return
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	informerFactory.Start(stopCh)

	informerFactory.WaitForCacheSync(stopCh)
}

func cleanUpOrchestratedPods(config *config.ZkOperatorConfig, namespace *corev1.Namespace) error {

	if !ContainsFinalizer(namespace, targetFinalizer) {
		//Finalizer is not present. No need to do any cleanup.
		return nil
	}

	// Remove mutating webhook configuration.
	err := DeleteMutatingWebhookConfiguration(config.Webhook.Name)
	if err != nil {
		logger.Error(FINALIZER_LOG_TAG, "Failed to delete mutating webhook configuration ", err)
		return err
	}

	err = RestartMarkedNamespacesIfNeeded(true)
	if err != nil {
		logger.Error(FINALIZER_LOG_TAG, "Failed to restart marked namespaces ", err)
		return err
	}

	// Remove finalizer from the namespace.
	RemoveFinalizer(namespace, targetFinalizer)

	return nil
}

func ContainsFinalizer(namespace *corev1.Namespace, finalizerName string) bool {
	for _, f := range namespace.Finalizers {
		if f == finalizerName {
			return true
		}
	}
	return false
}

func RemoveFinalizer(namespace *corev1.Namespace, finalizerName string) {
	finalizers := []string{}
	for _, f := range namespace.Finalizers {
		if f != finalizerName {
			finalizers = append(finalizers, f)
		}
	}
}
